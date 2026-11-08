// CON-TECH-02: จุดเริ่มต้นของเว็บแอป React (Vite) — ไม่เปิด StrictMode เพราะ AuthProvider ยิง GET /api/auth/session ทุกครั้งที่ mount
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router";
import { App } from "./App";

createRoot(document.getElementById("root")!).render(
  <BrowserRouter>
    <App />
  </BrowserRouter>,
);
