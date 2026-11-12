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

export const handlers = [loginHandler, sessionHandler, occupancyHandler];

// ใช้กับ server.use(...) เพื่อบังคับให้ตอบ 503 {error:"service_unavailable"} (ASM-05, AC-AUTH-07, 08)
const serviceUnavailable = () =>
  HttpResponse.json({ error: "service_unavailable" }, { status: 503 });

export const loginUnavailableHandler = http.post(LOGIN_PATH, serviceUnavailable);
export const sessionUnavailableHandler = http.get(SESSION_PATH, serviceUnavailable);

// ใช้กับ server.use(...) เพื่อจำลองฐานข้อมูลใช้ไม่ได้ (plan 007 D6: 503 {error:"unavailable"})
export const occupancyUnavailableHandler = http.get(OCCUPANCY_PATH, () =>
  HttpResponse.json({ error: "unavailable" }, { status: 503 }),
);
