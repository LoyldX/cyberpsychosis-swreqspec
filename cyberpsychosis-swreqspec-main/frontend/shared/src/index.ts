/**
 * Shared exports - Types, Auth utilities, and API client
 * Used by both mobile and web apps
 *
 * Auth functions implement FR-AUTH-01: Persistent session with refresh token
 * API client implements real backend integration
 */

export * from './types';
export {
  login,
  logout,
  refreshAccessToken,
  getStoredSession,
  saveSession,
  clearSession,
  isSessionValid,
  getAccessToken,
} from './hooks/useAuth';
export { apiClient, createApiClient } from './api/client';
