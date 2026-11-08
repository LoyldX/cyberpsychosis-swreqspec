// รองรับ FR-AUTH-01, CON-TECH-02: ต่อ AuthProvider กับเส้นทางของเว็บ (ไม่รวม Router เพื่อให้ test ครอบด้วย MemoryRouter ได้)
import { Navigate, Route, Routes } from "react-router";
import { AuthProvider } from "./features/auth/AuthProvider";
import { HomePage } from "./features/auth/HomePage";
import { LoginPage } from "./features/auth/LoginPage";
import { RequireAuth } from "./features/auth/RequireAuth";

// รองรับ FR-AUTH-01, ASM-07: /login = เข้าสู่ระบบ, / = หน้าหลักที่ต้องเข้าสู่ระบบ, path อื่น → /
export function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route
          path="/"
          element={
            <RequireAuth>
              <HomePage />
            </RequireAuth>
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AuthProvider>
  );
}
