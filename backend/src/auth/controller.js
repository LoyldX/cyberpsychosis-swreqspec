/**
 * Auth Controller
 * SPEC: 001-auth-session (FR-AUTH-01, NFR-REL-01, CON-CACHE-01)
 * Handles login, logout, and token refresh
 */

const bcrypt = require('bcryptjs');
const { PrismaClient } = require('@prisma/client');
const { generateAccessToken, generateRefreshToken, verifyRefreshToken } = require('../shared/jwt');
const { storeSession, deleteSession, storeRefreshToken, isRefreshTokenValid } = require('../shared/redis');

const prisma = new PrismaClient();

/**
 * Login: Authenticate member and issue tokens
 * FR-AUTH-01: Support member login with persistent session
 * NFR-REL-01: Use Redis + refresh token for session persistence
 *
 * ASSUMPTION Q1: Login via student_id (student ID); password hashed with bcrypt
 * ASSUMPTION Q2: Session TTL = 24 hours
 *
 * Request:
 *   POST /api/auth/login
 *   { student_id, password }
 *
 * Response:
 *   { access_token, refresh_token, expires_in, user: { id, student_id } }
 *
 * @param {object} req - Express request
 * @param {object} res - Express response
 */
async function login(req, res) {
  try {
    const { student_id: studentId, password } = req.body;

    // Validation
    if (!studentId || !password) {
      return res.status(400).json({
        error: 'Bad Request',
        message: 'Missing student_id or password',
      });
    }

    // Find member by student ID
    const member = await prisma.member.findUnique({
      where: { studentId },
    });

    if (!member) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Invalid credentials',
      });
    }

    // TODO: Verify password against hashed password in DB
    // For MVP, placeholder logic - team should implement password hashing
    const passwordValid = password === 'demo123'; // TEMPORARY FOR TESTING ONLY

    if (!passwordValid) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Invalid credentials',
      });
    }

    // Generate tokens
    const accessTokenData = generateAccessToken(member.id, member.studentId);
    const refreshTokenData = generateRefreshToken(member.id, member.studentId);

    // Store session in Redis (NFR-REL-01: CON-CACHE-01)
    await storeSession(member.id, accessTokenData.payload.jti, accessTokenData.payload);
    await storeRefreshToken(member.id, refreshTokenData.token, refreshTokenData.expiresIn);

    // Store refresh token in database for invalidation tracking
    await prisma.refreshToken.create({
      data: {
        memberId: member.id,
        token: refreshTokenData.token,
        expiresAt: new Date(Date.now() + refreshTokenData.expiresIn * 1000),
      },
    });

    res.json({
      access_token: accessTokenData.token,
      refresh_token: refreshTokenData.token,
      expires_in: accessTokenData.expiresIn,
      user: {
        id: member.id,
        student_id: member.studentId,
        name: member.name,
      },
    });
  } catch (err) {
    console.error('Login error:', err);
    res.status(500).json({
      error: 'Internal Server Error',
      message: 'Login failed',
    });
  }
}

/**
 * Refresh: Obtain new access token using refresh token
 * FR-AUTH-01: Allow members to stay logged in without re-entering credentials
 * NFR-REL-01: Prevent session dropout (Q2: max 0.1% unintended logout)
 *
 * Request:
 *   POST /api/auth/refresh
 *   { refresh_token }
 *
 * Response:
 *   { access_token, expires_in }
 *
 * @param {object} req - Express request
 * @param {object} res - Express response
 */
async function refresh(req, res) {
  try {
    const { refresh_token: refreshToken } = req.body;

    if (!refreshToken) {
      return res.status(400).json({
        error: 'Bad Request',
        message: 'Missing refresh_token',
      });
    }

    // Verify refresh token signature
    let payload;
    try {
      payload = verifyRefreshToken(refreshToken);
    } catch (err) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Invalid or expired refresh token',
      });
    }

    // Check if refresh token is still valid (not revoked)
    const isValid = await isRefreshTokenValid(payload.user_id, refreshToken);
    if (!isValid) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Refresh token has been revoked',
      });
    }

    // Verify token still exists in database (not revoked)
    const tokenRecord = await prisma.refreshToken.findUnique({
      where: { token: refreshToken },
    });

    if (!tokenRecord || tokenRecord.revokedAt) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Refresh token is invalid or revoked',
      });
    }

    // Get member info
    const member = await prisma.member.findUnique({
      where: { id: payload.user_id },
    });

    if (!member) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Member not found',
      });
    }

    // Generate new access token
    const newAccessTokenData = generateAccessToken(member.id, member.studentId);

    // Store new session in Redis
    await storeSession(member.id, newAccessTokenData.payload.jti, newAccessTokenData.payload);

    res.json({
      access_token: newAccessTokenData.token,
      expires_in: newAccessTokenData.expiresIn,
    });
  } catch (err) {
    console.error('Refresh token error:', err);
    res.status(500).json({
      error: 'Internal Server Error',
      message: 'Token refresh failed',
    });
  }
}

/**
 * Logout: Revoke tokens and delete session
 * FR-AUTH-01: Allow members to explicitly log out
 * NFR-REL-01: Clean up session from Redis
 *
 * Request:
 *   POST /api/auth/logout
 *   Headers: Authorization: Bearer <access_token>
 *
 * Response:
 *   { message: 'Logged out successfully' }
 *
 * @param {object} req - Express request (requires authenticateToken middleware)
 * @param {object} res - Express response
 */
async function logout(req, res) {
  try {
    if (!req.user) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'No active session to log out',
      });
    }

    const { id: memberId, tokenId } = req.user;

    // Delete session from Redis
    await deleteSession(memberId, tokenId);

    // Revoke refresh tokens in database
    await prisma.refreshToken.updateMany({
      where: { memberId, revokedAt: null },
      data: { revokedAt: new Date() },
    });

    res.json({
      message: 'Logged out successfully',
    });
  } catch (err) {
    console.error('Logout error:', err);
    res.status(500).json({
      error: 'Internal Server Error',
      message: 'Logout failed',
    });
  }
}

/**
 * Get current user info
 * FR-AUTH-01: Allow members to fetch their profile
 *
 * Request:
 *   GET /api/auth/me
 *   Headers: Authorization: Bearer <access_token>
 *
 * Response:
 *   { id, student_id, name, email, phone }
 *
 * @param {object} req - Express request (requires authenticateToken middleware)
 * @param {object} res - Express response
 */
async function getCurrentUser(req, res) {
  try {
    if (!req.user) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'No active session',
      });
    }

    const member = await prisma.member.findUnique({
      where: { id: req.user.id },
      select: {
        id: true,
        studentId: true,
        name: true,
        email: true,
        phone: true,
      },
    });

    if (!member) {
      return res.status(404).json({
        error: 'Not Found',
        message: 'Member not found',
      });
    }

    res.json(member);
  } catch (err) {
    console.error('Get user error:', err);
    res.status(500).json({
      error: 'Internal Server Error',
      message: 'Failed to retrieve user info',
    });
  }
}

module.exports = {
  login,
  refresh,
  logout,
  getCurrentUser,
};
