// รองรับ FR-AUTH-01, ASM-01, ASM-05, ASM-06, CON-TECH-02: หน้าเข้าสู่ระบบด้วยรหัสนักศึกษาและรหัสผ่าน (ไม่มีลงทะเบียน/ลืมรหัสผ่าน ตาม Out of scope)
import { useState, type FormEvent } from "react";
import { Navigate, useNavigate } from "react-router";
import { login } from "./api";
import { useAuth } from "./AuthProvider";

// ข้อความเดียวที่ไม่บอกว่าผิดที่รหัสนักศึกษาหรือรหัสผ่าน (AC-AUTH-06, ASM-06)
const INVALID_CREDENTIALS_MESSAGE = "รหัสนักศึกษาหรือรหัสผ่านไม่ถูกต้อง";
// ระบบ Redis/PostgreSQL ไม่ตอบหรือเครือข่ายล้ม (ASM-05)
const UNAVAILABLE_MESSAGE = "ระบบไม่พร้อมให้บริการ กรุณาลองใหม่อีกครั้ง";

// รองรับ FR-AUTH-01, ASM-05, ASM-06: ฟอร์มเข้าสู่ระบบ สำเร็จ → signIn แล้วไปหน้าหลัก, ผิด/ระบบไม่ตอบ → แสดงข้อความและคงฟอร์มไว้
export function LoginPage() {
  const { status, signIn } = useAuth();
  const navigate = useNavigate();
  const [studentId, setStudentId] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // เข้าสู่ระบบอยู่แล้ว → ไปหน้าหลัก (AC-AUTH-10: ไม่ต้องกรอกซ้ำ)
  if (status === "authenticated") return <Navigate to="/" replace />;
  // กำลังตรวจ session ตอนเปิดเว็บ ยังไม่แสดงฟอร์มเพื่อไม่ให้กรอกซ้ำโดยไม่จำเป็น
  if (status === "loading" && !submitting) return <p>กำลังตรวจสอบการเข้าสู่ระบบ...</p>;

  // รองรับ FR-AUTH-01: ส่งรหัสไป login() แล้วแยกผลตาม LoginResult
  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (submitting) return;
    setSubmitting(true);
    setError(null);
    const result = await login(studentId, password);
    if (result.kind === "ok") {
      await signIn(result.refreshToken);
      navigate("/", { replace: true });
      return;
    }
    setError(result.kind === "invalid_credentials" ? INVALID_CREDENTIALS_MESSAGE : UNAVAILABLE_MESSAGE);
    setSubmitting(false);
  }

  return (
    <form onSubmit={(event) => void handleSubmit(event)}>
      <h1>เข้าสู่ระบบ</h1>
      <div>
        <label htmlFor="student-id">รหัสนักศึกษา</label>
        <input
          id="student-id"
          name="student_id"
          type="text"
          autoComplete="username"
          value={studentId}
          onChange={(event) => setStudentId(event.target.value)}
        />
      </div>
      <div>
        <label htmlFor="password">รหัสผ่าน</label>
        <input
          id="password"
          name="password"
          type="password"
          autoComplete="current-password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
        />
      </div>
      {error && <p role="alert">{error}</p>}
      <button type="submit" disabled={submitting}>
        {submitting ? "กำลังเข้าสู่ระบบ..." : "เข้าสู่ระบบ"}
      </button>
    </form>
  );
}
