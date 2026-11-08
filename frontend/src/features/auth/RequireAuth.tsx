// รองรับ FR-AUTH-01, ASM-05, ASM-07: route guard — ไม่ได้เข้าสู่ระบบไป /login, ระบบไม่ตอบให้ลองใหม่โดยไม่ลบ token
import type { ReactNode } from "react";
import { Navigate } from "react-router";
import { useAuth } from "./AuthProvider";

// รองรับ FR-AUTH-01, ASM-05, ASM-07: ครอบหน้าที่ต้องเข้าสู่ระบบ
export function RequireAuth({ children }: { children: ReactNode }) {
  const { status, retry } = useAuth();

  if (status === "loading") return <p>กำลังตรวจสอบการเข้าสู่ระบบ...</p>;
  if (status === "unauthenticated") return <Navigate to="/login" replace />;
  if (status === "unavailable") {
    return (
      <div role="alert">
        <p>ระบบไม่พร้อมให้บริการในขณะนี้ กรุณาลองใหม่อีกครั้ง</p>
        <button type="button" onClick={() => void retry()}>
          ลองใหม่
        </button>
      </div>
    );
  }
  return <>{children}</>;
}
