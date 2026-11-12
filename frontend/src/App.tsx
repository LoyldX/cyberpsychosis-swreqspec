// รองรับ FR-AUTH-01, CON-TECH-02: ต่อ AuthProvider กับเส้นทางของเว็บ (ไม่รวม Router เพื่อให้ test ครอบด้วย MemoryRouter ได้)
import { Navigate, Route, Routes } from "react-router";
import { AuthProvider } from "./features/auth/AuthProvider";
import { HomePage } from "./features/auth/HomePage";
import { LoginPage } from "./features/auth/LoginPage";
import { RequireAuth } from "./features/auth/RequireAuth";
import { ClassBookingPage } from "./features/booking/ClassBookingPage";
import { InventoryCatalogPage } from "./features/inventory/InventoryCatalogPage";
import { OccupancyDashboardPage } from "./features/occupancy/OccupancyDashboardPage";
import { RosterPage } from "./features/roster/RosterPage";

// รองรับ FR-AUTH-01, FR-INV-01, CON-TECH-02: /inventory แสดง Catalog ภายใต้ route guard เดียวกับหน้าหลัก
// รองรับ FR-BKG-01, CON-TECH-02: /classes (จองคลาส) ต้องล็อกอินก่อนตาม precondition ของ 005 จึงครอบด้วย RequireAuth
// รองรับ FR-ROST-01, ASM-06 (004): /roster (จัดตารางงาน) sprint นี้เปิดได้โดยไม่ต้องล็อกอิน จึงไม่ครอบด้วย RequireAuth
// รองรับ FR-BI-01, ASM-07: /dashboard (Occupancy Dashboard) เปิดได้โดยไม่ต้องล็อกอิน จึงไม่ครอบด้วย RequireAuth
export function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/dashboard" element={<OccupancyDashboardPage />} />
        <Route path="/roster" element={<RosterPage />} />
        <Route
          path="/inventory"
          element={
            <RequireAuth>
              <InventoryCatalogPage />
            </RequireAuth>
          }
        />
        <Route
          path="/classes"
          element={
            <RequireAuth>
              <ClassBookingPage />
            </RequireAuth>
          }
        />
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
