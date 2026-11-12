// รองรับ FR-BKG-01, CON-TECH-02: ชนิดข้อมูลตามสัญญา API ใน specs/005-class-booking/plan.md ข้อ 4

// คลาสกลุ่มที่เตรียมไว้ล่วงหน้า (ASM-03, ASM-04) — เวลาเป็น ISO 8601
export type ClassSummary = {
  id: number;
  trainer_id: number;
  starts_at: string;
  ends_at: string;
  capacity: number;
};

// การจองหนึ่งรายการของสมาชิกที่ล็อกอินอยู่ (ASM-06)
export type Booking = {
  id: number;
  class_id: number;
  member_id: number;
  created_at: string;
};

// รหัสเหตุผลที่ backend ตอบเมื่อจองไม่สำเร็จ (ASM-05, DOM-BKG-01)
export type BookingRejection =
  | "invalid_request"
  | "class_not_found"
  | "class_started"
  | "already_booked"
  | "class_full"
  | "trainer_off_shift"
  | "trainer_busy";

export type ClassesResult =
  | { kind: "ok"; classes: ClassSummary[] }
  | { kind: "unauthorized" } // HTTP 401 จาก RequireSession ของ 001
  | { kind: "unavailable" }; // HTTP 503, สถานะอื่น หรือเครือข่ายล้ม (ASM-07)

export type CreateBookingResult =
  | { kind: "ok"; booking: Booking } // HTTP 201
  | { kind: "rejected"; reason: BookingRejection; message: string } // HTTP 400/404/409 (ASM-05)
  | { kind: "unauthorized" }
  | { kind: "unavailable" };

export type BookingResult =
  | { kind: "ok"; booking: Booking } // HTTP 200
  | { kind: "not_found" } // HTTP 404 ไม่พบหรือเป็นของคนอื่น (plan D5)
  | { kind: "unauthorized" }
  | { kind: "unavailable" };
