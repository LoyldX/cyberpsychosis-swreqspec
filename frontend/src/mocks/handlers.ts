// CON-TECH-02: API จำลองตามสัญญาใน specs/001-auth-session/plan.md §4.1 (FR-AUTH-01)
// ค่าด้านล่างเป็นข้อมูลจำลองสำหรับ test เท่านั้น ไม่ใช่ข้อมูลของผลิตภัณฑ์ (Q-03)
import { http, HttpResponse } from "msw";

export const FAKE_STUDENT_ID = "fake-student-0001";
export const FAKE_PASSWORD = "fake-password-for-tests";
export const FAKE_MEMBER_ID = 1;
export const FAKE_REFRESH_TOKEN = "fake-refresh-token-for-tests";
// 7 วัน = 604800 วินาที (ASM-02) ใช้เป็นค่า expires_in ของการจำลองเท่านั้น
export const FAKE_EXPIRES_IN = 604800;

const LOGIN_PATH = "/api/auth/login";
const SESSION_PATH = "/api/auth/session";

// POST /api/auth/login: 200 {refresh_token, expires_in} | 401 {error:"invalid_credentials"} (body เดียวกันทุกกรณีที่ไม่ถูกต้อง, AC-AUTH-06)
const loginHandler = http.post(LOGIN_PATH, async ({ request }) => {
  let body: unknown;
  try {
    body = await request.json();
  } catch {
    body = null;
  }
  const { student_id, password } = (body ?? {}) as Record<string, unknown>;
  if (student_id === FAKE_STUDENT_ID && password === FAKE_PASSWORD) {
    return HttpResponse.json({
      refresh_token: FAKE_REFRESH_TOKEN,
      expires_in: FAKE_EXPIRES_IN,
    });
  }
  return HttpResponse.json({ error: "invalid_credentials" }, { status: 401 });
});

// GET /api/auth/session: 200 {member:{id, student_id}} | 401 เมื่อไม่มี/ไม่ถูกต้อง (AC-AUTH-02, 04)
const sessionHandler = http.get(SESSION_PATH, ({ request }) => {
  const token = request.headers.get("Authorization")?.match(/^Bearer (.+)$/)?.[1];
  if (token === FAKE_REFRESH_TOKEN) {
    return HttpResponse.json({
      member: { id: FAKE_MEMBER_ID, student_id: FAKE_STUDENT_ID },
    });
  }
  return new HttpResponse(null, { status: 401 });
});

// GET /api/occupancy: 200 {count, as_of} ตามสัญญาใน specs/007-occupancy-dashboard/plan.md §4.1 (FR-BI-01)
// ค่าด้านล่างเป็นข้อมูลจำลองสำหรับ test เท่านั้น
export const FAKE_OCCUPANCY_COUNT = 3;
export const FAKE_OCCUPANCY_AS_OF = "2026-11-12T12:00:00+07:00";
const OCCUPANCY_PATH = "/api/occupancy";

const occupancyHandler = http.get(OCCUPANCY_PATH, () =>
  HttpResponse.json({ count: FAKE_OCCUPANCY_COUNT, as_of: FAKE_OCCUPANCY_AS_OF }),
);

// API การจองของ 005 ตามสัญญาใน specs/005-class-booking/plan.md ข้อ 4 (FR-BKG-01, FR-BKG-02)
// ทุก endpoint ต้องมี Bearer FAKE_REFRESH_TOKEN (RequireSession ของ 001) ค่าด้านล่างเป็นข้อมูลจำลองสำหรับ test เท่านั้น
export const FAKE_CLASSES = [
  { id: 11, trainer_id: 1, starts_at: "2026-11-13T10:00:00+07:00", ends_at: "2026-11-13T11:00:00+07:00", capacity: 10 },
  { id: 12, trainer_id: 1, starts_at: "2026-11-13T10:30:00+07:00", ends_at: "2026-11-13T11:30:00+07:00", capacity: 8 },
];
// คลาส 12 ทับกับคิวสอนอื่นของเทรนเนอร์ จึงจองไม่ได้ (409 trainer_busy, DOM-BKG-01)
export const FAKE_BUSY_CLASS_ID = 12;
export const FAKE_BOOKING_ID = 101;
export const FAKE_BOOKING_CREATED_AT = "2026-11-13T08:00:00+07:00";
export const FAKE_TRAINER_BUSY_MESSAGE = "จองไม่สำเร็จ เพราะเทรนเนอร์ไม่ว่าง มีคิวสอนคลาสอื่นในช่วงเวลานี้";
const CLASSES_PATH = "/api/classes";
const BOOKINGS_PATH = "/api/bookings";

const hasFakeSession = (request: Request) =>
  request.headers.get("Authorization") === `Bearer ${FAKE_REFRESH_TOKEN}`;
const unauthorized = () => HttpResponse.json({ error: "unauthorized" }, { status: 401 });
const fakeBooking = (classId: number) => ({
  id: FAKE_BOOKING_ID,
  class_id: classId,
  member_id: FAKE_MEMBER_ID,
  created_at: FAKE_BOOKING_CREATED_AT,
});
// คลาสที่ถูกจองล่าสุดใน mock เพื่อให้ GET /api/bookings/:id คืนการจองเดียวกับที่ POST สร้าง
let lastBookedClassId = FAKE_CLASSES[0].id;

// GET /api/classes: 200 [ClassSummary] | 401
const classesHandler = http.get(CLASSES_PATH, ({ request }) =>
  hasFakeSession(request) ? HttpResponse.json(FAKE_CLASSES) : unauthorized(),
);

// POST /api/bookings: 201 Booking | 404 class_not_found | 409 trainer_busy | 401
const createBookingHandler = http.post(BOOKINGS_PATH, async ({ request }) => {
  if (!hasFakeSession(request)) return unauthorized();
  const body = (await request.json().catch(() => null)) as { class_id?: number } | null;
  const classId = body?.class_id;
  if (classId === FAKE_BUSY_CLASS_ID) {
    return HttpResponse.json({ error: "trainer_busy", message: FAKE_TRAINER_BUSY_MESSAGE }, { status: 409 });
  }
  if (!FAKE_CLASSES.some((c) => c.id === classId)) {
    return HttpResponse.json({ error: "class_not_found", message: "ไม่พบคลาสนี้" }, { status: 404 });
  }
  lastBookedClassId = classId as number;
  return HttpResponse.json(fakeBooking(lastBookedClassId), { status: 201 });
});

// GET /api/bookings/:id: 200 Booking | 404 booking_not_found | 401
const bookingHandler = http.get(`${BOOKINGS_PATH}/:id`, ({ request, params }) => {
  if (!hasFakeSession(request)) return unauthorized();
  if (Number(params.id) !== FAKE_BOOKING_ID) {
    return HttpResponse.json({ error: "booking_not_found", message: "ไม่พบการจองนี้" }, { status: 404 });
  }
  return HttpResponse.json(fakeBooking(lastBookedClassId));
});

export const handlers = [
  loginHandler,
  sessionHandler,
  occupancyHandler,
  classesHandler,
  createBookingHandler,
  bookingHandler,
];

// ใช้กับ server.use(...) เพื่อบังคับให้ตอบ 503 {error:"service_unavailable"} (ASM-05, AC-AUTH-07, 08)
const serviceUnavailable = () =>
  HttpResponse.json({ error: "service_unavailable" }, { status: 503 });

export const loginUnavailableHandler = http.post(LOGIN_PATH, serviceUnavailable);
export const sessionUnavailableHandler = http.get(SESSION_PATH, serviceUnavailable);

// ใช้กับ server.use(...) เพื่อจำลองฐานข้อมูลใช้ไม่ได้ (plan 007 D6: 503 {error:"unavailable"})
export const occupancyUnavailableHandler = http.get(OCCUPANCY_PATH, () =>
  HttpResponse.json({ error: "unavailable" }, { status: 503 }),
);

// ใช้กับ server.use(...) เพื่อจำลองฐานข้อมูลไม่ตอบ (ASM-07: 503 {error:"unavailable", message})
const bookingUnavailable = () =>
  HttpResponse.json(
    { error: "unavailable", message: "ระบบไม่พร้อมให้บริการ กรุณาลองใหม่อีกครั้ง" },
    { status: 503 },
  );
export const classesUnavailableHandler = http.get(CLASSES_PATH, bookingUnavailable);
export const createBookingUnavailableHandler = http.post(BOOKINGS_PATH, bookingUnavailable);
