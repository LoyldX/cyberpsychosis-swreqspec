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

// API ตารางงานของ 004 ตามสัญญาใน specs/004-trainer-roster/plan.md ข้อ 4 (FR-ROST-01) — ไม่ต้องล็อกอิน (ASM-06)
// เก็บรายการไว้ในหน่วยความจำ ใช้ resetFakeRoster() ก่อนแต่ละ test; กฎใน mock เป็นแบบย่อของ backend (ทับซ้อน/เวลาสอนนอกเวลาเข้างาน)
export const FAKE_TRAINERS = [
  { id: 1, name: "ครูเอ" },
  { id: 2, name: "ครูบี" },
];
type FakeEntry = {
  id: number;
  trainer_id: number;
  kind: "work" | "teach";
  weekday: number;
  start_time: string;
  end_time: string;
  cancelled: boolean;
};
const initialRoster = (): FakeEntry[] => [
  { id: 1, trainer_id: 1, kind: "work", weekday: 1, start_time: "09:00", end_time: "12:00", cancelled: false },
];
let fakeRoster = initialRoster();
let nextRosterId = 100;
export function resetFakeRoster() {
  fakeRoster = initialRoster();
  nextRosterId = 100;
}
const TRAINERS_PATH = "/api/trainers";
const ROSTER_PATH = "/api/roster";
const rosterError = (status: number, error: string, message: string) =>
  HttpResponse.json({ error, message }, { status });
const toMinutes = (t: string) => Number(t.slice(0, 2)) * 60 + Number(t.slice(3, 5));
const validSlot = (b: Partial<FakeEntry>) =>
  typeof b.weekday === "number" && b.weekday >= 1 && b.weekday <= 7 &&
  /^\d\d:\d\d$/.test(b.start_time ?? "") && /^\d\d:\d\d$/.test(b.end_time ?? "") &&
  toMinutes(b.start_time as string) < toMinutes(b.end_time as string);
// ตรวจรายการที่ใช้งานของเทรนเนอร์ในวันเดียวกัน: ทับซ้อน (ต่อกันพอดีได้) แล้วเวลาสอนต้องอยู่ในเวลาเข้างานช่วงใดช่วงหนึ่ง
function checkFakeSchedule(entries: FakeEntry[]): "overlap" | "teach_outside_work" | null {
  const active = entries.filter((e) => !e.cancelled);
  for (const a of active) {
    for (const b of active) {
      if (a !== b && a.kind === b.kind && a.weekday === b.weekday &&
        toMinutes(a.start_time) < toMinutes(b.end_time) && toMinutes(b.start_time) < toMinutes(a.end_time)) {
        return "overlap";
      }
    }
  }
  for (const t of active.filter((e) => e.kind === "teach")) {
    const covered = active.some((w) => w.kind === "work" && w.weekday === t.weekday &&
      toMinutes(w.start_time) <= toMinutes(t.start_time) && toMinutes(t.end_time) <= toMinutes(w.end_time));
    if (!covered) return "teach_outside_work";
  }
  return null;
}
const scheduleMessages = {
  overlap: "บันทึกไม่ได้ เพราะเวลาทับซ้อนกับรายการอื่นของเทรนเนอร์คนนี้",
  teach_outside_work: "บันทึกไม่ได้ เพราะเวลาสอนต้องอยู่ภายในเวลาเข้างาน",
};
const trainerExists = (id: number) => FAKE_TRAINERS.some((t) => t.id === id);

// GET /api/trainers: 200 [Trainer]
const trainersHandler = http.get(TRAINERS_PATH, () => HttpResponse.json(FAKE_TRAINERS));

// GET /api/trainers/:id/roster: 200 [RosterEntry] | 404 trainer_not_found
const rosterHandler = http.get(`${TRAINERS_PATH}/:id/roster`, ({ params }) => {
  const id = Number(params.id);
  if (!trainerExists(id)) return rosterError(404, "trainer_not_found", "ไม่พบเทรนเนอร์นี้");
  return HttpResponse.json(fakeRoster.filter((e) => e.trainer_id === id));
});

// POST /api/trainers/:id/roster: 201 | 400 | 404 | 409
const createRosterHandler = http.post(`${TRAINERS_PATH}/:id/roster`, async ({ params, request }) => {
  const id = Number(params.id);
  if (!trainerExists(id)) return rosterError(404, "trainer_not_found", "ไม่พบเทรนเนอร์นี้");
  const body = ((await request.json().catch(() => null)) ?? {}) as Partial<FakeEntry>;
  if ((body.kind !== "work" && body.kind !== "teach") || !validSlot(body)) {
    return rosterError(400, "invalid_request", "ข้อมูลตารางงานไม่ถูกต้อง");
  }
  const entry: FakeEntry = { id: nextRosterId++, trainer_id: id, kind: body.kind, weekday: body.weekday as number,
    start_time: body.start_time as string, end_time: body.end_time as string, cancelled: false };
  const problem = checkFakeSchedule([...fakeRoster.filter((e) => e.trainer_id === id), entry]);
  if (problem) return rosterError(409, problem, scheduleMessages[problem]);
  fakeRoster.push(entry);
  return HttpResponse.json(entry, { status: 201 });
});

// PATCH /api/roster/:entryId: 200 | 400 | 404 | 409
const updateRosterHandler = http.patch(`${ROSTER_PATH}/:entryId`, async ({ params, request }) => {
  const old = fakeRoster.find((e) => e.id === Number(params.entryId));
  if (!old) return rosterError(404, "entry_not_found", "ไม่พบรายการนี้");
  const body = ((await request.json().catch(() => null)) ?? {}) as Partial<FakeEntry>;
  if (!validSlot(body)) return rosterError(400, "invalid_request", "ข้อมูลตารางงานไม่ถูกต้อง");
  const updated: FakeEntry = { ...old, weekday: body.weekday as number, start_time: body.start_time as string,
    end_time: body.end_time as string };
  const problem = checkFakeSchedule(fakeRoster.map((e) => (e.id === old.id ? updated : e)).filter((e) => e.trainer_id === old.trainer_id));
  if (problem) return rosterError(409, problem, scheduleMessages[problem]);
  fakeRoster = fakeRoster.map((e) => (e.id === old.id ? updated : e));
  return HttpResponse.json(updated);
});

// POST /api/roster/:entryId/cancel: 200 | 404
const cancelRosterHandler = http.post(`${ROSTER_PATH}/:entryId/cancel`, ({ params }) => {
  const old = fakeRoster.find((e) => e.id === Number(params.entryId));
  if (!old) return rosterError(404, "entry_not_found", "ไม่พบรายการนี้");
  const cancelled = { ...old, cancelled: true };
  fakeRoster = fakeRoster.map((e) => (e.id === old.id ? cancelled : e));
  return HttpResponse.json(cancelled);
});

// DELETE /api/roster/:entryId: 204 | 404
const deleteRosterHandler = http.delete(`${ROSTER_PATH}/:entryId`, ({ params }) => {
  const old = fakeRoster.find((e) => e.id === Number(params.entryId));
  if (!old) return rosterError(404, "entry_not_found", "ไม่พบรายการนี้");
  fakeRoster = fakeRoster.filter((e) => e.id !== old.id);
  return new HttpResponse(null, { status: 204 });
});

export const handlers = [
  loginHandler,
  sessionHandler,
  occupancyHandler,
  classesHandler,
  createBookingHandler,
  bookingHandler,
  trainersHandler,
  rosterHandler,
  createRosterHandler,
  updateRosterHandler,
  cancelRosterHandler,
  deleteRosterHandler,
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

// ใช้กับ server.use(...) เพื่อจำลองฐานข้อมูลไม่ตอบในหน้าตารางงาน (004 ASM-08)
const rosterUnavailable = () =>
  HttpResponse.json(
    { error: "unavailable", message: "ระบบไม่พร้อมให้บริการ กรุณาลองใหม่อีกครั้ง" },
    { status: 503 },
  );
export const trainersUnavailableHandler = http.get(TRAINERS_PATH, rosterUnavailable);
export const createRosterUnavailableHandler = http.post(`${TRAINERS_PATH}/:id/roster`, rosterUnavailable);
