// CON-TECH-02: ตั้งค่า test ฝั่ง React — เปิด/ปิด msw server รอบ test และโหลด jest-dom
import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterAll, afterEach, beforeAll } from "vitest";
import { server } from "../mocks/server";

// request ที่ไม่มี handler รองรับให้ error เพื่อไม่ให้ test ยิง API จริงโดยไม่รู้ตัว
beforeAll(() => server.listen({ onUnhandledFrame: "error" }));
afterEach(() => {
  cleanup();
  server.resetHandlers();
});
afterAll(() => server.close());
