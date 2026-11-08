// รองรับ FR-AUTH-01, CON-TECH-02: API client (fetch) ตามสัญญา plan.md §4.1 และที่เก็บ Refresh Token ฝั่งเว็บ (ASM-07)

export type Member = { id: number; student_id: string };

export type LoginResult =
  | { kind: "ok"; refreshToken: string; expiresIn: number }
  | { kind: "invalid_credentials" } // HTTP 401 (AC-AUTH-06)
  | { kind: "unavailable" }; // HTTP 503, สถานะอื่นที่ไม่ใช่ 200/401, หรือเครือข่ายล้ม (ASM-05)

export type SessionResult =
  | { kind: "ok"; member: Member }
  | { kind: "unauthorized" } // HTTP 401 (AC-AUTH-03, 04)
  | { kind: "unavailable" }; // HTTP 503, สถานะอื่นที่ไม่ใช่ 200/401, หรือเครือข่ายล้ม (ASM-05)

// ชื่อ key เดียวของ Refresh Token ใน localStorage (ASM-07: อยู่ข้ามการปิดเบราว์เซอร์)
export const REFRESH_TOKEN_KEY = "gym.refresh_token";

// รองรับ ASM-07: อ่าน token จาก localStorage อย่างปลอดภัย (คืน null ถ้าไม่มีหรือเข้าถึงไม่ได้)
export function readToken(): string | null {
  try {
    return window.localStorage.getItem(REFRESH_TOKEN_KEY);
  } catch {
    return null;
  }
}

// รองรับ ASM-07: เก็บ token ลง localStorage อย่างปลอดภัย
export function saveToken(token: string): void {
  try {
    window.localStorage.setItem(REFRESH_TOKEN_KEY, token);
  } catch {
    // เก็บไม่ได้: ไม่ throw ให้ UI พัง (สมาชิกอาจต้องเข้าสู่ระบบใหม่รอบหน้า)
  }
}

// รองรับ ASM-05/AC-AUTH-03: ลบ token (ใช้เมื่อ session ตอบ 401 เท่านั้น)
export function clearToken(): void {
  try {
    window.localStorage.removeItem(REFRESH_TOKEN_KEY);
  } catch {
    // ลบไม่ได้: ไม่ throw
  }
}

// รองรับ FR-AUTH-01, ASM-01, ASM-06: POST /api/auth/login
export async function login(studentId: string, password: string): Promise<LoginResult> {
  try {
    const res = await fetch("/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ student_id: studentId, password }),
    });
    if (res.status === 401) return { kind: "invalid_credentials" };
    if (!res.ok) return { kind: "unavailable" };
    const body = (await res.json()) as { refresh_token?: unknown; expires_in?: unknown };
    if (typeof body.refresh_token !== "string" || typeof body.expires_in !== "number") {
      return { kind: "unavailable" };
    }
    return { kind: "ok", refreshToken: body.refresh_token, expiresIn: body.expires_in };
  } catch {
    return { kind: "unavailable" };
  }
}

// รองรับ FR-AUTH-01, ASM-02, ASM-05: GET /api/auth/session ด้วย Bearer Refresh Token
export async function fetchSession(token: string): Promise<SessionResult> {
  try {
    const res = await fetch("/api/auth/session", {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.status === 401) return { kind: "unauthorized" };
    if (!res.ok) return { kind: "unavailable" };
    const body = (await res.json()) as { member?: { id?: unknown; student_id?: unknown } };
    const m = body.member;
    if (!m || typeof m.id !== "number" || typeof m.student_id !== "string") {
      return { kind: "unavailable" };
    }
    return { kind: "ok", member: { id: m.id, student_id: m.student_id } };
  } catch {
    return { kind: "unavailable" };
  }
}
