/**
 * useAuth Utilities - Manages FR-AUTH-01 persistent session with refresh token
 * References: FR-AUTH-01, NFR-REL-01, UI-AUTH-01
 *
 * Access token TTL: Backend returns expires_in: 900 seconds (15 minutes)
 * Refresh token: Stored in localStorage; can refresh multiple times within 24h
 * Q2 Resolved: Backend provides expires_in with each response
 */

import { SessionState, User, AuthResponse } from '../types';
import { apiClient } from '../api/client';

/**
 * Storage keys for persistent session (FR-AUTH-01)
 * Used to restore session on app/page reload (UI-AUTH-01)
 */
const STORAGE_KEYS = {
  ACCESS_TOKEN: 'gym_access_token',
  REFRESH_TOKEN: 'gym_refresh_token',
  USER: 'gym_user',
  EXPIRES_AT: 'gym_expires_at',
};

/**
 * Get current stored session
 * Used to check if user is already authenticated on app launch
 */
export function getStoredSession(): SessionState {
  if (typeof window === 'undefined') {
    return {
      user: null,
      access_token: null,
      refresh_token: null,
      expiresAt: null,
      isLoading: false,
      error: null,
    };
  }

  try {
    return {
      user: JSON.parse(localStorage.getItem(STORAGE_KEYS.USER) || 'null'),
      access_token: localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN),
      refresh_token: localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN),
      expiresAt: localStorage.getItem(STORAGE_KEYS.EXPIRES_AT) ?
        parseInt(localStorage.getItem(STORAGE_KEYS.EXPIRES_AT)!) : null,
      isLoading: false,
      error: null,
    };
  } catch {
    // Corrupted storage - clear it
    Object.values(STORAGE_KEYS).forEach(key => localStorage.removeItem(key));
    return {
      user: null,
      access_token: null,
      refresh_token: null,
      expiresAt: null,
      isLoading: false,
      error: null,
    };
  }
}

/**
 * Save session to persistent storage (FR-AUTH-01)
 * Called after successful login or token refresh
 */
export function saveSession(user: User, accessToken: string, refreshToken: string, expiresIn: number): void {
  if (typeof window === 'undefined') return;

  const expiresAt = Date.now() + (expiresIn * 1000);

  localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(user));
  localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, accessToken);
  localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, refreshToken);
  localStorage.setItem(STORAGE_KEYS.EXPIRES_AT, expiresAt.toString());
}

/**
 * Clear session from storage (called on logout)
 */
export function clearSession(): void {
  if (typeof window === 'undefined') return;
  Object.values(STORAGE_KEYS).forEach(key => localStorage.removeItem(key));
}

/**
 * Check if stored session is still valid
 * UI-AUTH-01: App launch checks this; if valid, skip login screen
 */
export function isSessionValid(): boolean {
  const session = getStoredSession();
  if (!session.access_token || !session.refresh_token || !session.expiresAt) {
    return false;
  }
  return Date.now() < session.expiresAt;
}

/**
 * Login with student_id + password
 * ASSUMPTION: Auth uses student_id + password (matches backend contract)
 * Returns true if successful, throws Error if failed
 */
export async function login(student_id: string, password: string): Promise<User> {
  const response = await apiClient.login(student_id, password);

  // FR-AUTH-01: Save tokens to storage
  // Backend returns expires_in: 900 seconds (15 minutes for access token)
  saveSession(response.user, response.access_token, response.refresh_token, 900);

  return response.user;
}

/**
 * Refresh access token using stored refresh token
 * FR-AUTH-01: Called when access token expires (or before)
 */
export async function refreshAccessToken(): Promise<{ accessToken: string; expiresIn: number } | null> {
  const session = getStoredSession();

  if (!session.refresh_token) {
    return null;
  }

  try {
    const response = await apiClient.refreshToken(session.refresh_token);

    // Update access token in storage (keep same refresh token)
    if (session.user) {
      saveSession(session.user, response.access_token, session.refresh_token, response.expires_in);
    }

    return {
      accessToken: response.access_token,
      expiresIn: response.expires_in,
    };
  } catch (error) {
    // Refresh failed - clear session and require re-login
    clearSession();
    throw error;
  }
}

/**
 * Logout: Clear session and call backend logout
 */
export async function logout(): Promise<void> {
  try {
    await apiClient.logout();
  } finally {
    clearSession();
  }
}

/**
 * Get current access token for API requests
 * Used by API client to attach Authorization header
 */
export function getAccessToken(): string | null {
  const session = getStoredSession();

  // Check if token is still valid
  if (!session.access_token || !session.expiresAt) {
    return null;
  }

  if (Date.now() > session.expiresAt) {
    // Token expired - return null (caller should refresh)
    return null;
  }

  return session.access_token;
}
