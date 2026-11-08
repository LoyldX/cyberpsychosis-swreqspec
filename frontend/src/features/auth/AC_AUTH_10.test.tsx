// ตรวจ AC-AUTH-10 (FR-AUTH-01, ASM-07) และพฤติกรรม LoginPage/HomePage (AC-AUTH-01, 06 ฝั่งหน้าจอ, ASM-05, ASM-06) ด้วย msw
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it } from "vitest";
import { App } from "../../App";
import {
  FAKE_PASSWORD,
  FAKE_REFRESH_TOKEN,
  FAKE_STUDENT_ID,
  loginUnavailableHandler,
} from "../../mocks/handlers";
import { server } from "../../mocks/server";
import { REFRESH_TOKEN_KEY } from "./api";

function renderApp(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <App />
    </MemoryRouter>,
  );
}

function fillAndSubmit(studentId: string, password: string) {
  fireEvent.change(screen.getByLabelText("รหัสนักศึกษา"), { target: { value: studentId } });
  fireEvent.change(screen.getByLabelText("รหัสผ่าน"), { target: { value: password } });
  fireEvent.click(screen.getByRole("button", { name: "เข้าสู่ระบบ" }));
}

beforeEach(() => {
  window.localStorage.clear();
});

describe("AC-AUTH-10", () => {
  it("AC_AUTH_10: resume after reload", async () => {
    // Given: สมาชิกเข้าสู่ระบบไว้ก่อนหน้า (token ยังอยู่ใน localStorage ภายในอายุ Session)
    window.localStorage.setItem(REFRESH_TOKEN_KEY, FAKE_REFRESH_TOKEN);

    // เปิดเว็บครั้งแรกแล้วปิด (unmount) เพื่อจำลองปิดเบราว์เซอร์ โดย localStorage ยังอยู่
    const first = renderApp("/");
    expect(await screen.findByText(FAKE_STUDENT_ID)).toBeInTheDocument();
    first.unmount();

    // When: เปิดเว็บใหม่ (App ตัวใหม่ทั้งหมด ไม่มี state เดิมค้าง)
    renderApp("/");

    // Then: เห็นรหัสนักศึกษาบน HomePage ไม่เห็นฟอร์มเข้าสู่ระบบ ไม่มีปุ่มออกจากระบบ
    expect(await screen.findByText(FAKE_STUDENT_ID)).toBeInTheDocument();
    expect(screen.queryByLabelText("รหัสนักศึกษา")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("รหัสผ่าน")).not.toBeInTheDocument();
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
    expect(screen.queryByText(/ออกจากระบบ/)).not.toBeInTheDocument();
  });
});

describe("LoginPage และ HomePage", () => {
  it("signs in with valid credentials, lands on HomePage with the student id and saves the token", async () => {
    renderApp("/login");
    fillAndSubmit(FAKE_STUDENT_ID, FAKE_PASSWORD);

    expect(await screen.findByText(FAKE_STUDENT_ID)).toBeInTheDocument();
    expect(screen.queryByLabelText("รหัสผ่าน")).not.toBeInTheDocument();
    expect(window.localStorage.getItem(REFRESH_TOKEN_KEY)).toBe(FAKE_REFRESH_TOKEN);
  });

  it("shows one generic error for wrong credentials and saves no token", async () => {
    renderApp("/login");
    fillAndSubmit(FAKE_STUDENT_ID, "wrong-password");

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("รหัสนักศึกษาหรือรหัสผ่านไม่ถูกต้อง");
    expect(screen.getByLabelText("รหัสนักศึกษา")).toBeInTheDocument(); // ยังอยู่ที่ฟอร์ม
    expect(window.localStorage.getItem(REFRESH_TOKEN_KEY)).toBeNull();

    // รหัสนักศึกษาที่ไม่มีในระบบต้องได้ข้อความเดียวกันทุกตัวอักษร (AC-AUTH-06)
    const firstMessage = alert.textContent;
    fillAndSubmit("unknown-student", "wrong-password");
    await waitFor(() => expect(screen.getByRole("alert").textContent).toBe(firstMessage));
    expect(window.localStorage.getItem(REFRESH_TOKEN_KEY)).toBeNull();
  });

  it("shows the unavailable message, keeps the form and saves no token when login gets 503", async () => {
    server.use(loginUnavailableHandler);
    renderApp("/login");
    fillAndSubmit(FAKE_STUDENT_ID, FAKE_PASSWORD);

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "ระบบไม่พร้อมให้บริการ กรุณาลองใหม่อีกครั้ง",
    );
    expect(screen.getByLabelText("รหัสนักศึกษา")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "เข้าสู่ระบบ" })).toBeEnabled();
    expect(window.localStorage.getItem(REFRESH_TOKEN_KEY)).toBeNull();
  });

  it("redirects to the login form when there is no token and the path is /", async () => {
    renderApp("/");

    expect(await screen.findByLabelText("รหัสนักศึกษา")).toBeInTheDocument();
    expect(screen.getByLabelText("รหัสผ่าน")).toHaveAttribute("type", "password");
    expect(screen.queryByText(FAKE_STUDENT_ID)).not.toBeInTheDocument();
  });

  it("redirects an already signed-in member from /login to HomePage", async () => {
    window.localStorage.setItem(REFRESH_TOKEN_KEY, FAKE_REFRESH_TOKEN);
    renderApp("/login");

    expect(await screen.findByText(FAKE_STUDENT_ID)).toBeInTheDocument();
    expect(screen.queryByLabelText("รหัสนักศึกษา")).not.toBeInTheDocument();
  });

  it("disables the submit button while the login request is in flight", async () => {
    renderApp("/login");
    fillAndSubmit(FAKE_STUDENT_ID, FAKE_PASSWORD);

    expect(screen.getByRole("button")).toBeDisabled();
    expect(await screen.findByText(FAKE_STUDENT_ID)).toBeInTheDocument();
  });
});
