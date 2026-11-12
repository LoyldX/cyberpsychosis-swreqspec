// ตรวจหน้าจัดตารางงาน (FR-ROST-01, CON-TECH-02) กับ API จำลอง (msw) ตามสัญญา plan.md ข้อ 4
import { fireEvent, render, screen, within } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it } from "vitest";
import { App } from "../../App";
import { createRosterUnavailableHandler, resetFakeRoster } from "../../mocks/handlers";
import { server } from "../../mocks/server";
import { RosterPage, UNAVAILABLE_MESSAGE } from "./RosterPage";

beforeEach(() => {
  resetFakeRoster();
  window.localStorage.clear(); // ไม่มี token ใด ๆ (ASM-06)
});

async function chooseTrainer(name: string) {
  const select = screen.getByLabelText("เทรนเนอร์");
  await screen.findByRole("option", { name });
  const option = screen.getByRole("option", { name }) as HTMLOptionElement;
  fireEvent.change(select, { target: { value: option.value } });
}

function fillNew(kind: "work" | "teach", weekday: string, start: string, end: string) {
  const form = screen.getByRole("form", { name: "เพิ่มรายการ" });
  fireEvent.change(within(form).getByLabelText("ประเภท"), { target: { value: kind } });
  fireEvent.change(within(form).getByLabelText("วัน"), { target: { value: weekday } });
  fireEvent.change(within(form).getByLabelText("เวลาเริ่ม"), { target: { value: start } });
  fireEvent.change(within(form).getByLabelText("เวลาจบ"), { target: { value: end } });
  return form;
}

describe("RosterPage", () => {
  // AC-ROST-01: Given เทรนเนอร์ที่ยังไม่มีเวลาเข้างาน When กำหนดเวลาเข้างานให้ Then เวลาเข้างานถูกบันทึกและเรียกดูได้
  it("AC_ROST_01 เพิ่มเวลาเข้างานแล้วเห็นในตาราง", async () => {
    render(<RosterPage />);
    await chooseTrainer("ครูบี");
    expect(await screen.findByText("ยังไม่มีตารางงาน")).toBeInTheDocument();

    const form = fillNew("work", "3", "13:00", "17:00");
    fireEvent.click(within(form).getByRole("button", { name: "เพิ่ม" }));

    const list = await screen.findByRole("list", { name: "ตารางงาน" });
    expect(list).toHaveTextContent("พุธ 13:00–17:00 เวลาเข้างาน");
  });

  // AC-ROST-03: Given เทรนเนอร์ที่มีตารางอยู่แล้ว When ปรับปรุงเวลาเข้างาน Then เมื่อเรียกดูซ้ำเห็นค่าใหม่
  it("AC_ROST_03 แก้เวลาเข้างานแล้วเห็นค่าใหม่", async () => {
    render(<RosterPage />);
    await chooseTrainer("ครูเอ");
    const list = await screen.findByRole("list", { name: "ตารางงาน" });
    expect(list).toHaveTextContent("จันทร์ 09:00–12:00 เวลาเข้างาน");

    fireEvent.click(within(list).getByRole("button", { name: "แก้ไข" }));
    const form = screen.getByRole("form", { name: "แก้ไขรายการ" });
    fireEvent.change(within(form).getByLabelText("เวลาจบ"), { target: { value: "13:00" } });
    fireEvent.click(within(form).getByRole("button", { name: "บันทึกการแก้ไข" }));

    expect(await screen.findByText(/จันทร์ 09:00–13:00 เวลาเข้างาน/)).toBeInTheDocument();
    expect(screen.queryByRole("form", { name: "แก้ไขรายการ" })).not.toBeInTheDocument();
  });

  it("409 แสดงเหตุผลจาก backend (ASM-02)", async () => {
    render(<RosterPage />);
    await chooseTrainer("ครูเอ");
    await screen.findByRole("list", { name: "ตารางงาน" });
    const form = fillNew("work", "1", "11:00", "13:00");
    fireEvent.click(within(form).getByRole("button", { name: "เพิ่ม" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("ทับซ้อน");
  });

  it("503 แจ้งให้ลองใหม่และคงค่าในฟอร์ม (ASM-08)", async () => {
    server.use(createRosterUnavailableHandler);
    render(<RosterPage />);
    await chooseTrainer("ครูบี");
    const form = fillNew("work", "2", "10:00", "11:00");
    fireEvent.click(within(form).getByRole("button", { name: "เพิ่ม" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(UNAVAILABLE_MESSAGE);
    expect(within(form).getByLabelText("วัน")).toHaveValue("2");
    expect(within(form).getByLabelText("เวลาเริ่ม")).toHaveValue("10:00");
    expect(within(form).getByLabelText("เวลาจบ")).toHaveValue("11:00");
  });

  it("ยกเลิกแล้วยังเห็นพร้อมสถานะ และลบแล้วหายไป (ASM-01)", async () => {
    render(<RosterPage />);
    await chooseTrainer("ครูเอ");
    const list = await screen.findByRole("list", { name: "ตารางงาน" });
    fireEvent.click(within(list).getByRole("button", { name: "ยกเลิก" }));
    expect(await screen.findByText(/\(ยกเลิกแล้ว\)/)).toBeInTheDocument();

    fireEvent.click(within(screen.getByRole("list", { name: "ตารางงาน" })).getByRole("button", { name: "ลบ" }));
    expect(await screen.findByText("ยังไม่มีตารางงาน")).toBeInTheDocument();
  });
});

describe("route /roster", () => {
  it("เปิดได้โดยไม่ต้องล็อกอิน (ASM-06)", async () => {
    render(
      <MemoryRouter initialEntries={["/roster"]}>
        <App />
      </MemoryRouter>,
    );
    expect(await screen.findByRole("heading", { name: "จัดตารางงานเทรนเนอร์" })).toBeInTheDocument();
    expect(screen.queryByLabelText("รหัสนักศึกษา")).not.toBeInTheDocument();
  });
});
