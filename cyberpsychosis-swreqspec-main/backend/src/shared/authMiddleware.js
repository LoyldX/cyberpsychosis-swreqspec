/**
 * Authentication Middleware
 * SPEC: 001-auth-session (FR-AUTH-01, NFR-REL-01)
 * Validates JWT access token and session in Redis
 */

const { verifyAccessToken } = require('./jwt');
const { getSession } = require('./redis');

/**
 * Middleware: Verify JWT and session (protected routes)
 * FR-AUTH-01: Protect API endpoints with valid access token
 * NFR-REL-01: Verify session exists in Redis
 *
 * @param {object} req - Express request
 * @param {object} res - Express response
 * @param {function} next - Express next middleware
 */
async function authenticateToken(req, res, next) {
  try {
    const authHeader = req.headers.authorization;
    const token = authHeader && authHeader.split(' ')[1]; // "Bearer <token>"

    if (!token) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Missing access token in Authorization header',
      });
    }

    // Verify JWT signature
    let payload;
    try {
      payload = verifyAccessToken(token);
    } catch (err) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: err.message,
      });
    }

    // Verify session exists in Redis (NFR-REL-01)
    const session = await getSession(payload.user_id, payload.jti);
    if (!session) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Session expired or revoked',
      });
    }

    // Attach user info to request for downstream handlers
    req.user = {
      id: payload.user_id,
      studentId: payload.student_id,
      tokenId: payload.jti,
      type: payload.type,
    };

    next();
  } catch (err) {
    console.error('Auth middleware error:', err);
    res.status(500).json({
      error: 'Internal Server Error',
      message: 'Authentication check failed',
    });
  }
}

/**
 * Middleware: Optional authentication (attach user if token present, otherwise continue)
 * Useful for endpoints that work with or without auth
 *
 * @param {object} req - Express request
 * @param {object} res - Express response
 * @param {function} next - Express next middleware
 */
async function optionalAuth(req, res, next) {
  try {
    const authHeader = req.headers.authorization;
    const token = authHeader && authHeader.split(' ')[1];

    if (token) {
      try {
        const payload = verifyAccessToken(token);
        const session = await getSession(payload.user_id, payload.jti);

        if (session) {
          req.user = {
            id: payload.user_id,
            studentId: payload.student_id,
            tokenId: payload.jti,
            type: payload.type,
          };
        }
      } catch (err) {
        // Token invalid/expired, but that's okay for optional auth
        console.debug('Optional auth token invalid:', err.message);
      }
    }

    next();
  } catch (err) {
    console.error('Optional auth middleware error:', err);
    next(); // Continue regardless of error
  }
}

/**
 * Middleware: Verify user ID matches request parameter
 * Prevents users from accessing other users' data
 *
 * @param {string} paramName - URL parameter name (default: 'memberId')
 * @returns {function} Middleware function
 */
function validateOwnResource(paramName = 'memberId') {
  return (req, res, next) => {
    if (!req.user) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Authentication required',
      });
    }

    const resourceId = parseInt(req.params[paramName], 10);
    if (req.user.id !== resourceId) {
      return res.status(403).json({
        error: 'Forbidden',
        message: 'You do not have permission to access this resource',
      });
    }

    next();
  };
}

module.exports = {
  authenticateToken,
  optionalAuth,
  validateOwnResource,
};
