/**
 * Redis Client Configuration
 * SPEC: 001-auth-session (CON-CACHE-01: Use Redis for session management)
 * NFR-REL-01: Session persistence using Redis with TTL-based expiry
 *
 * Session key pattern: user:{user_id}:session:{token_id}
 * Payload format: { user_id, student_id, iat, exp, type: 'access'|'refresh' }
 */

const redis = require('redis');

const REDIS_URL = process.env.REDIS_URL || 'redis://localhost:6379';
const SESSION_TTL_HOURS = parseInt(process.env.SESSION_TTL_HOURS || '24', 10);
const SESSION_TTL_SECONDS = SESSION_TTL_HOURS * 3600;

let redisClient = null;

/**
 * Get or create Redis client singleton
 * @returns {Promise<redis.RedisClient>}
 */
async function getRedisClient() {
  if (redisClient) {
    return redisClient;
  }

  try {
    redisClient = redis.createClient({ url: REDIS_URL });

    redisClient.on('error', (err) => {
      console.error('Redis Client Error:', err);
    });

    redisClient.on('connect', () => {
      console.log('✓ Redis connected');
    });

    redisClient.on('disconnect', () => {
      console.log('✗ Redis disconnected');
    });

    await redisClient.connect();
    return redisClient;
  } catch (error) {
    console.error('Failed to connect Redis:', error);
    throw error;
  }
}

/**
 * Store session in Redis
 * FR-AUTH-01: Session persistence via Redis
 * NFR-REL-01: Q2 assumption: 24-hour TTL, max 0.1% unintended dropout
 *
 * @param {number} userId - Member ID
 * @param {string} tokenId - Unique token identifier
 * @param {object} payload - Session payload { user_id, student_id, iat, exp, type }
 * @returns {Promise<void>}
 */
async function storeSession(userId, tokenId, payload) {
  const client = await getRedisClient();
  const key = `user:${userId}:session:${tokenId}`;

  await client.setEx(
    key,
    SESSION_TTL_SECONDS,
    JSON.stringify(payload),
  );
}

/**
 * Retrieve session from Redis
 * FR-AUTH-01: Validate ongoing session
 *
 * @param {number} userId - Member ID
 * @param {string} tokenId - Token identifier
 * @returns {Promise<object|null>} Parsed session payload or null if expired/not found
 */
async function getSession(userId, tokenId) {
  const client = await getRedisClient();
  const key = `user:${userId}:session:${tokenId}`;

  const data = await client.get(key);
  if (!data) {
    return null;
  }

  try {
    return JSON.parse(data);
  } catch (err) {
    console.error(`Failed to parse session ${key}:`, err);
    return null;
  }
}

/**
 * Delete session from Redis (logout)
 * FR-AUTH-01: Explicit session termination on logout
 *
 * @param {number} userId - Member ID
 * @param {string} tokenId - Token identifier
 * @returns {Promise<void>}
 */
async function deleteSession(userId, tokenId) {
  const client = await getRedisClient();
  const key = `user:${userId}:session:${tokenId}`;

  await client.del(key);
}

/**
 * Check if session exists
 * @param {number} userId - Member ID
 * @param {string} tokenId - Token identifier
 * @returns {Promise<boolean>}
 */
async function sessionExists(userId, tokenId) {
  const client = await getRedisClient();
  const key = `user:${userId}:session:${tokenId}`;

  const exists = await client.exists(key);
  return exists === 1;
}

/**
 * Store refresh token for invalidation tracking
 * Used to prevent reuse after logout
 *
 * @param {number} userId - Member ID
 * @param {string} token - JWT refresh token string
 * @param {number} expirySeconds - Token expiry in seconds
 * @returns {Promise<void>}
 */
async function storeRefreshToken(userId, token, expirySeconds) {
  const client = await getRedisClient();
  const key = `user:${userId}:refresh_token:${token}`;

  await client.setEx(key, expirySeconds, 'valid');
}

/**
 * Check if refresh token is valid (not revoked)
 * @param {number} userId - Member ID
 * @param {string} token - JWT refresh token string
 * @returns {Promise<boolean>}
 */
async function isRefreshTokenValid(userId, token) {
  const client = await getRedisClient();
  const key = `user:${userId}:refresh_token:${token}`;

  const exists = await client.exists(key);
  return exists === 1;
}

/**
 * Revoke refresh token (on logout)
 * @param {number} userId - Member ID
 * @param {string} token - JWT refresh token string
 * @returns {Promise<void>}
 */
async function revokeRefreshToken(userId, token) {
  const client = await getRedisClient();
  const key = `user:${userId}:refresh_token:${token}`;

  await client.del(key);
}

/**
 * Close Redis connection
 * @returns {Promise<void>}
 */
async function closeRedis() {
  if (redisClient) {
    await redisClient.quit();
    redisClient = null;
  }
}

module.exports = {
  getRedisClient,
  storeSession,
  getSession,
  deleteSession,
  sessionExists,
  storeRefreshToken,
  isRefreshTokenValid,
  revokeRefreshToken,
  closeRedis,
};
