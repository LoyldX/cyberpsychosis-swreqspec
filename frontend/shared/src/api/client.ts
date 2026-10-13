/**
 * API Client - Shared HTTP client for mobile and web
 * Backend API contract (PostgreSQL per Q4 assumption)
 *
 * Implements real backend integration:
 * - FR-AUTH-01: Token management and refresh
 * - FR-BKG-01, FR-BKG-02: Booking with trainer availability validation (DOM-BKG-01)
 * - Request/response mapping
 * - Error handling
 */

import { User, AuthResponse, Booking, InventoryItem, OccupancyData, StockMovement } from '../types';

const API_BASE_URL = typeof window !== 'undefined'
  ? window.location.origin + '/api'
  : 'http://localhost:3000/api';

interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
}

interface RefreshResponse {
  access_token: string;
  expires_in: number;
}

interface BookingListItem {
  id: string;
  memberId: string;
  classId: string;
  status: 'confirmed' | 'cancelled';
  createdAt: string;
  class: {
    id: string;
    trainerId: string;
    scheduledAt: string;
    durationMins: number;
    capacity: number;
    trainer: {
      id: string;
      name: string;
    };
  };
}

export interface ApiClient {
  // FR-AUTH-01: Authentication endpoints
  login: (student_id: string, password: string) => Promise<AuthResponse>;
  refreshToken: (refresh_token: string) => Promise<{ access_token: string; expires_in: number }>;
  logout: () => Promise<void>;
  getCurrentUser: () => Promise<User>;

  // FR-BKG-01, FR-BKG-02: Booking endpoints
  getMyBookings: (memberId: string) => Promise<Booking[]>;
  bookClass: (class_id: string) => Promise<Booking>;
  unbookClass: (booking_id: string) => Promise<void>;

  // FR-INV-01, FR-INV-02, FR-INV-03: Inventory endpoints (Phase 2)
  getInventoryItems: () => Promise<InventoryItem[]>;
  createInventoryItem: (item: Partial<InventoryItem>) => Promise<InventoryItem>;
  updateInventoryItem: (id: string, item: Partial<InventoryItem>) => Promise<InventoryItem>;
  deleteInventoryItem: (id: string) => Promise<void>;
  recordStockMovement: (movement: Partial<StockMovement>) => Promise<StockMovement>;

  // FR-BI-01: Dashboard/Occupancy endpoints (Phase 2)
  getOccupancyData: () => Promise<OccupancyData>;
}

/**
 * Create API client with real backend integration
 * Handles:
 * - Access token attachment to headers (FR-AUTH-01)
 * - Error responses mapping
 * - Token refresh on 401 (handled by auth hook)
 */
export function createApiClient(getAccessToken?: () => string | null): ApiClient {
  const getHeaders = () => {
    const token = getAccessToken?.();
    return {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
    };
  };

  const request = async <T>(
    method: string,
    endpoint: string,
    body?: unknown,
  ): Promise<T> => {
    const url = `${API_BASE_URL}${endpoint}`;

    try {
      const options: RequestInit = {
        method,
        headers: getHeaders() as HeadersInit,
        ...(body && { body: JSON.stringify(body) }),
      };

      const response = await fetch(url, options);

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        const error = new Error(
          errorData.message || `HTTP ${response.status}: ${response.statusText}`,
        ) as any;
        error.status = response.status;
        error.data = errorData;
        throw error;
      }

      return response.json();
    } catch (error) {
      if (error instanceof Error) {
        throw error;
      }
      throw new Error(`API request failed: ${error}`);
    }
  };

  return {
    // FR-AUTH-01: Authentication
    login: async (student_id: string, password: string): Promise<AuthResponse> => {
      const response = await request<LoginResponse>('POST', '/auth/login', {
        student_id,
        password,
      });

      return {
        access_token: response.access_token,
        refresh_token: response.refresh_token,
        user: response.user,
      };
    },

    refreshToken: async (refresh_token: string) => {
      return request<RefreshResponse>('POST', '/auth/refresh', { refresh_token });
    },

    logout: async () => {
      await request('POST', '/auth/logout');
    },

    getCurrentUser: async (): Promise<User> => {
      return request<User>('GET', '/auth/me');
    },

    // FR-BKG-01, FR-BKG-02: Booking
    getMyBookings: async (memberId: string): Promise<Booking[]> => {
      const response = await request<BookingListItem[]>('GET', `/booking/member/${memberId}`);

      // Transform backend response to frontend Booking type
      return response.map((item) => ({
        id: item.id,
        class_id: item.classId,
        member_id: item.memberId,
        status: item.status,
        created_at: item.createdAt,
      }));
    },

    bookClass: async (class_id: string): Promise<Booking> => {
      return request<Booking>('POST', '/booking', { class_id });
    },

    unbookClass: async (booking_id: string): Promise<void> => {
      await request('DELETE', `/booking/${booking_id}`);
    },

    // Phase 2: Inventory endpoints
    getInventoryItems: async () => {
      return request<InventoryItem[]>('GET', '/inventory');
    },

    createInventoryItem: async (item: Partial<InventoryItem>) => {
      return request<InventoryItem>('POST', '/inventory', item);
    },

    updateInventoryItem: async (id: string, item: Partial<InventoryItem>) => {
      return request<InventoryItem>('PATCH', `/inventory/${id}`, item);
    },

    deleteInventoryItem: async (id: string) => {
      await request('DELETE', `/inventory/${id}`);
    },

    recordStockMovement: async (movement: Partial<StockMovement>) => {
      return request<StockMovement>('POST', '/inventory/movements', movement);
    },

    getOccupancyData: async () => {
      return request<OccupancyData>('GET', '/occupancy');
    },
  };
}

export const apiClient = createApiClient();
