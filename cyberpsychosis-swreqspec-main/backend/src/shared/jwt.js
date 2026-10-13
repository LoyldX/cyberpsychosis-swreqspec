/**
 * JWT Token Management
 * SPEC: 001-auth-session (FR-AUTH-01, NFR-REL-01)
 * Q1 ASSUMPTION: JWT payload = { user_id, student_id, iat, exp, type: 'access'|'refresh' }
 */

const jwt = require('jsonwebtoken');
const { v4: uuidv4 } = require('uuid');

const JWT_ACCESS_SECRET = process.env.JWT_ACCESS_SECRET || 'dev-access-secret';
const JWT_REFRESH_SECRET = process.env.JWT_REFRESH_SECRET || 'dev-refresh-secret';
const JWT_ACCESS_EXPIRY = process.env.JWT_ACCESS_EXPIRY || '15m';
const JWT_REFRESH_EXPIRY = process.env.JWT_REFRESH_EXPIRY || '7d';

/**
 * Generate access token
 * FR-AUTH-01: Issue short-lived access token for API requests
 *
 * @param {number} memberId - Member/user ID
 * @param {string} studentId - Student ID
 * @returns {object} { token, expiresIn (seconds), payload }
 */
function generateAccessToken(memberId, studentId) {
  const payload = {
    user_id: memberId,
    student_id: studentId,
    type: 'access',
  };

  const token = jwt.sign(payload, JWT_ACCESS_SECRET, {
    expiresIn: JWT_ACCESS_EXPIRY,
    jti: uuidv4(), // Unique token ID for revocation tracking
  });

  // Decode to get expiry time
  const decoded = jwt.decode(token);

  return {
    token,
    expiresIn: decoded.exp - decoded.iat,
    payload: decoded,
  };
}

/**
 * Generate refresh token
 * FR-AUTH-01: Issue long-lived refresh token to obtain new access tokens
 * NFR-REL-01: Refresh token enables persistent session (no frequent re-login)
 *
 * @param {number} memberId - Member/user ID
 * @param {string} studentId - Student ID
 * @returns {object} { token, expiresIn (seconds), payload }
 */
function generateRefreshToken(memberId, studentId) {
  const payload = {
    user_id: memberId,
    student_id: studentId,
    type: 'refresh',
  };

  const token = jwt.sign(payload, JWT_REFRESH_SECRET, {
    expiresIn: JWT_REFRESH_EXPIRY,
    jti: uuidv4(),
  });

  const decoded = jwt.decode(token);

  return {
    token,
    expiresIn: decoded.exp - decoded.iat,
    payload: decoded,
  };
}

/**
 * Verify access token
 * FR-AUTH-01: Validate incoming API requests
 *
 * @param {string} token - JWT access token
 * @returns {object} Decoded payload if valid
 * @throws {Error} If token is invalid or expired
 */
function verifyAccessToken(token) {
  try {
    return jwt.verify(token, JWT_ACCESS_SECRET);
  } catch (err) {
    throw new Error(`Access token verification failed: ${err.message}`);
  }
}

/**
 * Verify refresh token
 * FR-AUTH-01: Validate refresh token for obtaining new access token
 *
 * @param {string} token - JWT refresh token
 * @returns {object} Decoded payload if valid
 * @throws {Error} If token is invalid or expired
 */
function verifyRefreshToken(token) {
  try {
    return jwt.verify(token, JWT_REFRESH_SECRET);
  } catch (err) {
    throw new Error(`Refresh token verification failed: ${err.message}`);
  }
}

/**
 * Decode token without verification (for troubleshooting)
 * @param {string} token - JWT token
 * @returns {object|null} Decoded payload or null if invalid
 */
function decodeToken(token) {
  try {
    return jwt.decode(token);
  } catch (err) {
    return null;
  }
}

module.exports = {
  generateAccessToken,
  generateRefreshToken,
  verifyAccessToken,
  verifyRefreshToken,
  decodeToken,
};
