// รองรับ FR-BKG-01, FR-BKG-02, CON-TECH-02: เรียก API การจองตามสัญญา plan.md ข้อ 4 พร้อม token ของ 001
import { readToken } from "../auth/api";
import type {
  Booking,
  BookingRejection,
  BookingResult,
  ClassSummary,
  ClassesResult,
  CreateBookingResult,
} from "./types";

export const CLASSES_URL = "/api/classes";
export const BOOKINGS_URL = "/api/bookings";

const REJECTIONS: readonly BookingRejection[] = [
  "invalid_request",
  "class_not_found",
  "class_started",
  "already_booked",
  "class_full",
  "trainer_off_shift",
  "trainer_busy",
];

// header Authorization จาก Refresh Token ของ 001 (ทุก endpoint ของ 005 อยู่หลัง RequireSession)
function authHeaders(): Record<string, string> {
  const token = readToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

// ส่งคำขอ ถ้าเครือข่ายล้มคืน null (ถือว่า unavailable ตาม ASM-07)
async function send(url: string, init: RequestInit = {}): Promise<Response | null> {
  try {
    return await fetch(url, { ...init, headers: { ...authHeaders(), ...init.headers } });
  } catch {
    return null;
  }
}

// รองรับ FR-BKG-01: GET /api/classes คลาสที่ยังไม่ถึงเวลาเริ่มให้สมาชิกเลือก (plan D4)
export async function fetchClasses(): Promise<ClassesResult> {
  const response = await send(CLASSES_URL);
  if (response?.status === 200) {
    return { kind: "ok", classes: (await response.json()) as ClassSummary[] };
  }
  return response?.status === 401 ? { kind: "unauthorized" } : { kind: "unavailable" };
}

// รองรับ FR-BKG-01, FR-BKG-02, DOM-BKG-01: POST /api/bookings จองคลาส แปลงรหัสเหตุผลที่ backend ส่งกลับ (ASM-05)
export async function createBooking(classId: number): Promise<CreateBookingResult> {
  const response = await send(BOOKINGS_URL, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ class_id: classId }),
  });
  if (!response) return { kind: "unavailable" };
  if (response.status === 201) {
    return { kind: "ok", booking: (await response.json()) as Booking };
  }
  if (response.status === 401) return { kind: "unauthorized" };
  if ([400, 404, 409].includes(response.status)) {
    const body = (await response.json().catch(() => null)) as { error?: string; message?: string } | null;
    const reason = REJECTIONS.find((r) => r === body?.error);
    if (reason) return { kind: "rejected", reason, message: body?.message ?? "" };
  }
  return { kind: "unavailable" };
}

// รองรับ FR-BKG-01 (AC-BKG-01 "เรียกดูการจองนั้นได้"): GET /api/bookings/:id ทีละรายการของตนเอง (plan D5)
export async function fetchBooking(id: number): Promise<BookingResult> {
  const response = await send(`${BOOKINGS_URL}/${id}`);
  if (response?.status === 200) {
    return { kind: "ok", booking: (await response.json()) as Booking };
  }
  if (response?.status === 404) return { kind: "not_found" };
  return response?.status === 401 ? { kind: "unauthorized" } : { kind: "unavailable" };
}
