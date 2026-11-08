// รองรับ CON-TECH-02: เว็บแอป React (Vite) อย่างเดียว; test ฝั่ง React ใช้ vitest + jsdom (000-shared §2)
import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
  },
});
