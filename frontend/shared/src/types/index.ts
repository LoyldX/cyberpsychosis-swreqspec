/**
 * Shared types for Gym Management System
 * Backend contracts referenced by spec IDs
 */

// FR-AUTH-01: Member authentication types
export interface LoginRequest {
  /** ASSUMPTION: Member auth uses student_id + password (Q2 TBD: verify with backend) */
  student_id: string;
  password: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string; // FR-AUTH-01: Refresh token for persistent session
  user: User;
}

export interface User {
  id: string;
  student_id: string;
  name: string;
  email?: string;
  user_type: 'member' | 'trainer' | 'staff' | 'manager'; // Q7: Assuming students only for MVP
}

export interface SessionState {
  user: User | null;
  access_token: string | null;
  refresh_token: string | null;
  // TODO: Q2 - Update TTL from backend config (currently assuming 24 hours)
  expiresAt: number | null;
  isLoading: boolean;
  error: string | null;
}

// FR-BKG-01: Booking types
export interface Class {
  id: string;
  trainer_id: string;
  trainer_name: string;
  start_time: string; // ISO 8601
  end_time: string;
  class_type: string;
  capacity: number;
  booked_count: number;
}

export interface Booking {
  id: string;
  class_id: string;
  member_id: string;
  status: 'confirmed' | 'cancelled';
  created_at: string;
}

// FR-BKG-02: DOM-BKG-01 - Trainer availability
export interface TrainerSchedule {
  trainer_id: string;
  date: string;
  available_slots: TimeSlot[];
  off_shift_times: TimeSlot[];
}

export interface TimeSlot {
  start_time: string;
  end_time: string;
}

// FR-INV-01, FR-INV-02, FR-INV-03: Inventory types
export interface InventoryItem {
  id: string;
  name: string;
  category: string;
  quantity: number;
  reorder_point: number; // DOM-INV-02
  unit: string;
}

export interface StockMovement {
  id: string;
  item_id: string;
  movement_type: 'issue' | 'return' | 'receive';
  quantity: number;
  responsible_person: string; // DOM-INV-01
  purpose?: string;
  return_date?: string;
  created_at: string;
}

// FR-BI-01: Dashboard types
export interface OccupancyData {
  current_count: number;
  timestamp: string;
  // DOM-CAP-01: Calculated from entry/exit signals (+1/-1)
}

// API Error response
export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: ApiError;
}
