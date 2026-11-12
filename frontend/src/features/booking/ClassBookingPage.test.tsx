// ตรวจหน้าจองคลาส (FR-BKG-01, FR-BKG-02, CON-TECH-02) กับ API จำลอง (msw) ตามสัญญา plan.md ข้อ 4
import { fireEvent, render, screen, within } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { App } from "../../App";
import {
  FAKE_BOOKING_ID,
  FAKE_BUSY_CLASS_ID,
  FAKE_CLASSES,
  FAKE_REFRESH_TOKEN,
  FAKE_TRAINER_BUSY_MESSAGE,
  classesUnavailableHandler,
  createBookingUnavailableHandler,
} from "../../mocks/handlers";
import { server } from "../../mocks/server";
import { clearToken, saveToken } from "../auth/api";
import { ClassBookingPage, UNAVAILABLE_MESSAGE } from "./ClassBookingPage";

beforeEach(() => saveToken(FAKE_REFRESH_TOKEN));
afterEach(() => clearToken());

async function bookButtons() {
  const list = await screen.findByRole("list", { name: "คลาสที่เปิดให้จอง" });
  return within(list).getAllByRole("button", { name: /^จองคลาส/ });
}

describe("ClassBookingPage", () => {
  // AC-BKG-01: Given สมาชิกที่ล็อกอินแล้ว และคลาสที่เปิดรับ When สมาชิกจองคลาสนั้น Then การจองสำเร็จและเรียกดูการจองนั้นได้
  it("AC_BKG_01 เห็นคลาส กดจอง แล้วแสดงการจองที่ได้", async () => {
    render(<ClassBookingPage />);
    const buttons = await bookButtons();
    expect(buttons).toHaveLength(FAKE_CLASSES.length);

    fireEvent.click(buttons[0]);

    const result = await screen.findByRole("status", { name: "ผลการจอง" });
    expect(result).toHaveTextContent("จองสำเร็จ");
    expect(result).toHaveTextContent(`หมายเลขการจอง ${FAKE_BOOKING_ID}`);
    expect(result).toHaveTextContent(`เทรนเนอร์ #${FAKE_CLASSES[0].trainer_id}`);
  });

  it("409 แสดงเหตุผลจาก backend และไม่แสดงว่าสำเร็จ (DOM-BKG-01, ASM-05)", async () => {
    render(<ClassBookingPage />);
    const buttons = await bookButtons();
    const busyIndex = FAKE_CLASSES.findIndex((c) => c.id === FAKE_BUSY_CLASS_ID);

    fireEvent.click(buttons[busyIndex]);

    expect(await screen.findByRole("alert")).toHaveTextContent(FAKE_TRAINER_BUSY_MESSAGE);
    expect(screen.queryByRole("status", { name: "ผลการจอง" })).not.toBeInTheDocument();
  });

  it("503 ตอนจองแสดงข้อความให้ลองใหม่ (ASM-07)", async () => {
    server.use(createBookingUnavailableHandler);
    render(<ClassBookingPage />);
    fireEvent.click((await bookButtons())[0]);
    expect(await screen.findByRole("alert")).toHaveTextContent(UNAVAILABLE_MESSAGE);
  });

  it("503 ตอนโหลดรายการคลาสแสดงข้อความให้ลองใหม่ (ASM-07)", async () => {
    server.use(classesUnavailableHandler);
    render(<ClassBookingPage />);
    expect(await screen.findByRole("alert")).toHaveTextContent(UNAVAILABLE_MESSAGE);
  });
});

describe("route /classes", () => {
  function renderApp() {
    return render(
      <MemoryRouter initialEntries={["/classes"]}>
        <App />
      </MemoryRouter>,
    );
  }

  it("สมาชิกที่ล็อกอินแล้วเปิดหน้าจองได้", async () => {
    renderApp();
    expect(await screen.findByRole("heading", { name: "จองคลาสออกกำลังกาย" })).toBeInTheDocument();
  });

  it("ยังไม่ล็อกอินถูกส่งไปหน้าเข้าสู่ระบบ (RequireAuth ของ 001)", async () => {
    clearToken();
    renderApp();
    expect(await screen.findByLabelText("รหัสนักศึกษา")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "จองคลาสออกกำลังกาย" })).not.toBeInTheDocument();
  });
});
