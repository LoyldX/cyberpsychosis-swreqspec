// ตรวจว่า msw ตอบตามสัญญา plan.md §4.1 (CON-TECH-02 โครง test; ไม่มี AC ตรง ๆ)
import { describe, expect, it } from "vitest";
import { server } from "./server";
import {
  FAKE_EXPIRES_IN,
  FAKE_MEMBER_ID,
  FAKE_PASSWORD,
  FAKE_REFRESH_TOKEN,
  FAKE_STUDENT_ID,
  loginUnavailableHandler,
  sessionUnavailableHandler,
} from "./handlers";

const login = (body: unknown) =>
  fetch("/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });

const session = (authorization?: string) =>
  fetch("/api/auth/session", {
    headers: authorization ? { Authorization: authorization } : {},
  });

describe("POST /api/auth/login (msw)", () => {
  it("returns 200 with refresh_token and expires_in for the fake valid student", async () => {
    const res = await login({ student_id: FAKE_STUDENT_ID, password: FAKE_PASSWORD });
    expect(res.status).toBe(200);
    expect(await res.json()).toEqual({
      refresh_token: FAKE_REFRESH_TOKEN,
      expires_in: FAKE_EXPIRES_IN,
    });
  });

  it("returns 401 invalid_credentials with the identical body for unknown student_id and wrong password", async () => {
    const unknownStudent = await login({ student_id: "no-such-student", password: FAKE_PASSWORD });
    const wrongPassword = await login({ student_id: FAKE_STUDENT_ID, password: "wrong" });
    expect(unknownStudent.status).toBe(401);
    expect(wrongPassword.status).toBe(401);
    const a = await unknownStudent.text();
    const b = await wrongPassword.text();
    expect(a).toBe(b);
    expect(JSON.parse(a)).toEqual({ error: "invalid_credentials" });
  });

  it("returns the same 401 body for missing fields", async () => {
    const res = await login({});
    expect(res.status).toBe(401);
    expect(await res.json()).toEqual({ error: "invalid_credentials" });
  });

  it("returns 503 service_unavailable when overridden with server.use", async () => {
    server.use(loginUnavailableHandler);
    const res = await login({ student_id: FAKE_STUDENT_ID, password: FAKE_PASSWORD });
    expect(res.status).toBe(503);
    expect(await res.json()).toEqual({ error: "service_unavailable" });
  });
});

describe("GET /api/auth/session (msw)", () => {
  it("returns 200 with member for the fake valid Bearer token", async () => {
    const res = await session(`Bearer ${FAKE_REFRESH_TOKEN}`);
    expect(res.status).toBe(200);
    expect(await res.json()).toEqual({
      member: { id: FAKE_MEMBER_ID, student_id: FAKE_STUDENT_ID },
    });
  });

  it("returns 401 when the Authorization header is missing", async () => {
    expect((await session()).status).toBe(401);
  });

  it("returns 401 for an invalid or malformed token", async () => {
    expect((await session("Bearer not-a-real-token")).status).toBe(401);
    expect((await session(FAKE_REFRESH_TOKEN)).status).toBe(401);
  });

  it("returns 503 service_unavailable when overridden with server.use", async () => {
    server.use(sessionUnavailableHandler);
    const res = await session(`Bearer ${FAKE_REFRESH_TOKEN}`);
    expect(res.status).toBe(503);
    expect(await res.json()).toEqual({ error: "service_unavailable" });
  });

  it("resets the 503 override after each test (valid token works again)", async () => {
    const res = await session(`Bearer ${FAKE_REFRESH_TOKEN}`);
    expect(res.status).toBe(200);
  });
});
