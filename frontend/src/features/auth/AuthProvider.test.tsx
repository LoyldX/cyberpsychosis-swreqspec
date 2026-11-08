// ตรวจ AuthProvider/RequireAuth (FR-AUTH-01, ASM-05, ASM-07, CON-TECH-02) ไม่มี AC ตรง ๆ จึงตั้งชื่อ test แบบอธิบาย
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { MemoryRouter, Route, Routes } from "react-router";
import { beforeEach, describe, expect, it } from "vitest";
import {
  FAKE_REFRESH_TOKEN,
  FAKE_STUDENT_ID,
  sessionUnavailableHandler,
} from "../../mocks/handlers";
import { server } from "../../mocks/server";
import { REFRESH_TOKEN_KEY } from "./api";
import { AuthProvider, useAuth } from "./AuthProvider";
import { RequireAuth } from "./RequireAuth";

// แสดงสถานะ/สมาชิกจาก useAuth เป็นข้อความเพื่อให้ assert ได้
function StatusProbe() {
  const { status, member } = useAuth();
  return (
    <div>
      <span data-testid="status">{status}</span>
      <span data-testid="member">{member ? `${member.id}:${member.student_id}` : "none"}</span>
    </div>
  );
}

function renderProbe() {
  return render(
    <AuthProvider>
      <StatusProbe />
    </AuthProvider>,
  );
}

function renderGuarded() {
  return render(
    <MemoryRouter initialEntries={["/"]}>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<div>หน้าเข้าสู่ระบบ</div>} />
          <Route
            path="/"
            element={
              <RequireAuth>
                <div>หน้าหลักที่ป้องกัน</div>
              </RequireAuth>
            }
          />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  );
}

const SESSION_URL = "/api/auth/session";
const networkErrorHandler = http.get(SESSION_URL, () => HttpResponse.error());

beforeEach(() => {
  window.localStorage.clear();
});

describe("AuthProvider", () => {
  it("is unauthenticated without calling the session endpoint when no token is stored", async () => {
    let sessionCalls = 0;
    server.use(
      http.get(SESSION_URL, () => {
        sessionCalls++;
        return new HttpResponse(null, { status: 401 });
      }),
    );
    renderProbe();
    await waitFor(() => expect(screen.getByTestId("status")).toHaveTextContent("unauthenticated"));
    expect(sessionCalls).toBe(0);
  });

  it("is authenticated with the member when the stored token gets 200", async () => {
    window.localStorage.setItem(REFRESH_TOKEN_KEY, FAKE_REFRESH_TOKEN);
    renderProbe();
    await waitFor(() => expect(screen.getByTestId("status")).toHaveTextContent("authenticated"));
    expect(screen.getByTestId("member")).toHaveTextContent(`1:${FAKE_STUDENT_ID}`);
    expect(window.localStorage.getItem(REFRESH_TOKEN_KEY)).toBe(FAKE_REFRESH_TOKEN);
  });

  it("removes the token and becomes unauthenticated when the session gets 401", async () => {
    window.localStorage.setItem(REFRESH_TOKEN_KEY, "expired-or-invalid-token");
    renderProbe();
    await waitFor(() => expect(screen.getByTestId("status")).toHaveTextContent("unauthenticated"));
    expect(window.localStorage.getItem(REFRESH_TOKEN_KEY)).toBeNull();
  });

  it("is unavailable and keeps the token when the session gets 503", async () => {
    window.localStorage.setItem(REFRESH_TOKEN_KEY, FAKE_REFRESH_TOKEN);
    server.use(sessionUnavailableHandler);
    renderProbe();
    await waitFor(() => expect(screen.getByTestId("status")).toHaveTextContent("unavailable"));
    expect(window.localStorage.getItem(REFRESH_TOKEN_KEY)).toBe(FAKE_REFRESH_TOKEN);
  });

  it("is unavailable and keeps the token on a network error", async () => {
    window.localStorage.setItem(REFRESH_TOKEN_KEY, FAKE_REFRESH_TOKEN);
    server.use(networkErrorHandler);
    renderProbe();
    await waitFor(() => expect(screen.getByTestId("status")).toHaveTextContent("unavailable"));
    expect(window.localStorage.getItem(REFRESH_TOKEN_KEY)).toBe(FAKE_REFRESH_TOKEN);
  });

  it("signIn stores the token and re-checks the session", async () => {
    function SignInProbe() {
      const { status, signIn } = useAuth();
      return (
        <div>
          <span data-testid="status">{status}</span>
          <button type="button" onClick={() => void signIn(FAKE_REFRESH_TOKEN)}>
            sign-in
          </button>
        </div>
      );
    }
    render(
      <AuthProvider>
        <SignInProbe />
      </AuthProvider>,
    );
    await waitFor(() => expect(screen.getByTestId("status")).toHaveTextContent("unauthenticated"));
    fireEvent.click(screen.getByText("sign-in"));
    await waitFor(() => expect(screen.getByTestId("status")).toHaveTextContent("authenticated"));
    expect(window.localStorage.getItem(REFRESH_TOKEN_KEY)).toBe(FAKE_REFRESH_TOKEN);
  });
});

describe("RequireAuth", () => {
  it("redirects to /login when unauthenticated", async () => {
    renderGuarded();
    expect(await screen.findByText("หน้าเข้าสู่ระบบ")).toBeInTheDocument();
    expect(screen.queryByText("หน้าหลักที่ป้องกัน")).not.toBeInTheDocument();
  });

  it("renders children when authenticated", async () => {
    window.localStorage.setItem(REFRESH_TOKEN_KEY, FAKE_REFRESH_TOKEN);
    renderGuarded();
    expect(await screen.findByText("หน้าหลักที่ป้องกัน")).toBeInTheDocument();
    expect(screen.queryByText("หน้าเข้าสู่ระบบ")).not.toBeInTheDocument();
  });

  it("shows a retry button when unavailable (no redirect, token kept) and retry works after the server recovers", async () => {
    window.localStorage.setItem(REFRESH_TOKEN_KEY, FAKE_REFRESH_TOKEN);
    server.use(sessionUnavailableHandler);
    renderGuarded();

    const retryButton = await screen.findByRole("button", { name: "ลองใหม่" });
    expect(screen.queryByText("หน้าเข้าสู่ระบบ")).not.toBeInTheDocument();
    expect(window.localStorage.getItem(REFRESH_TOKEN_KEY)).toBe(FAKE_REFRESH_TOKEN);

    server.resetHandlers(); // server กลับมาตอบปกติ
    fireEvent.click(retryButton);

    expect(await screen.findByText("หน้าหลักที่ป้องกัน")).toBeInTheDocument();
    expect(window.localStorage.getItem(REFRESH_TOKEN_KEY)).toBe(FAKE_REFRESH_TOKEN);
  });
});
