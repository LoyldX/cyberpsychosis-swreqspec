// รองรับ FR-AUTH-01, ASM-07, UI-AUTH-01 (Candidate): หน้าหลักแสดงรหัสนักศึกษาของสมาชิกที่ยืนยันแล้ว ไม่มีปุ่มออกจากระบบ (ASM-03)
import { useAuth } from "./AuthProvider";

// รองรับ FR-AUTH-01, ASM-07: แสดงรหัสนักศึกษาเพื่อให้เห็นว่า Session ถูกใช้งานต่อ (ใช้ใต้ RequireAuth เท่านั้น)
export function HomePage() {
  const { member } = useAuth();
  return (
    <p>
      รหัสนักศึกษา: <strong>{member?.student_id}</strong>
    </p>
  );
}
