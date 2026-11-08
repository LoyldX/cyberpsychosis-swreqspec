// CON-TECH-02: msw server สำหรับ test (node) ใช้ handler ตามสัญญา API ใน plan.md §4.1
import { setupServer } from "msw/node";
import { handlers } from "./handlers";

export const server = setupServer(...handlers);
