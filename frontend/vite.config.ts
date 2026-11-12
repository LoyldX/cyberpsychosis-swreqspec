// รองรับ CON-TECH-02: เว็บแอป React (Vite) อย่างเดียว; test ฝั่ง React ใช้ vitest + jsdom (000-shared §2)
import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

// dev proxy: ให้หน้าเว็บเรียก /api ของ backend Go ได้ตอน `npm run dev` (ทีมเลือกเอง ไม่ได้มาจาก spec)
// ปลายทางตรงกับ defaultHTTPAddr ":8080" ใน backend/internal/config/config.go
const apiProxyTarget = "http://localhost:8080";

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // ครอบทุก /api รวม SSE ของ 007 (/api/occupancy/stream) (FR-BI-01, NFR-PERF-02, CON-TECH-02)
      "/api": { target: apiProxyTarget, changeOrigin: true },
    },
  },
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
  },
});
