/**
 * Auth Routes
 * SPEC: 001-auth-session (FR-AUTH-01, NFR-REL-01, CON-CACHE-01)
 * Endpoints: login, refresh, logout, getCurrentUser
 */

const express = require('express');
const { authenticateToken } = require('../shared/authMiddleware');
const {
  login,
  refresh,
  logout,
  getCurrentUser,
} = require('./controller');

const router = express.Router();

/**
 * POST /api/auth/login
 * FR-AUTH-01: Member login
 * NFR-REL-01: Issue access & refresh tokens
 *
 * Request: { student_id, password }
 * Response: { access_token, refresh_token, expires_in, user }
 */
router.post('/login', async (req, res) => {
  await login(req, res);
});

/**
 * POST /api/auth/refresh
 * FR-AUTH-01: Obtain new access token
 * NFR-REL-01: Refresh token enables persistent session
 *
 * Request: { refresh_token }
 * Response: { access_token, expires_in }
 */
router.post('/refresh', async (req, res) => {
  await refresh(req, res);
});

/**
 * POST /api/auth/logout
 * FR-AUTH-01: Log out and revoke tokens
 *
 * Request: {} (Authorization header required)
 * Response: { message }
 */
router.post('/logout', authenticateToken, async (req, res) => {
  await logout(req, res);
});

/**
 * GET /api/auth/me
 * FR-AUTH-01: Fetch current user profile
 *
 * Response: { id, student_id, name, email, phone }
 */
router.get('/me', authenticateToken, async (req, res) => {
  await getCurrentUser(req, res);
});

module.exports = router;
