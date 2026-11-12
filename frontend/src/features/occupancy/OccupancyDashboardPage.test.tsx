// ตรวจ AC-BI-01 และ AC-BI-04 ฝั่งหน้าจอ (FR-BI-01, NFR-PERF-02, ASM-07) ด้วย API จำลอง (msw) และ EventSource จำลอง
import { act, render, screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { App } from "../../App";
import { FAKE_OCCUPANCY_AS_OF, occupancyUnavailableHandler } from "../../mocks/handlers";
import { server } from "../../mocks/server";
import { OCCUPANCY_EVENT, OCCUPANCY_STREAM_URL, OCCUPANCY_URL } from "./api";

/** EventSource จำลองแทนของเบราว์เซอร์ (jsdom ไม่มี) — เก็บทุกตัวที่ถูกสร้างไว้ให้ test สั่ง emit */
class FakeEventSource {
  static instances: FakeEventSource[] = [];
  listeners = new Map<string, ((event: MessageEvent) => void)[]>();
  closed = false;
  constructor(public url: string) {
    FakeEventSource.instances.push(this);
  }
  addEventListener(type: string, listener: (event: MessageEvent) => void) {
    this.listeners.set(type, [...(this.listeners.get(type) ?? []), listener]);
  }
  close() {
    this.closed = true;
  }
  emit(type: string, data: unknown) {
    for (const listener of this.listeners.get(type) ?? []) {
      listener(new MessageEvent(type, { data: JSON.stringify(data) }));
    }
  }
}

function renderDashboard() {
  return render(
    <MemoryRouter initialEntries={["/dashboard"]}>
      <App />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  window.localStorage.clear(); // ไม่มี Refresh Token = ยังไม่ได้ล็อกอิน (ASM-07)
  FakeEventSource.instances = [];
  vi.stubGlobal("EventSource", FakeEventSource);
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("OccupancyDashboardPage", () => {
  it("AC_BI_01: Given วันนี้รับสัญญาณเข้า 5 ออก 2 When ผู้จัดการเปิด Dashboard Then เห็นจำนวนปัจจุบันเป็น 3 (ไม่ต้องล็อกอิน)", async () => {
    // Given: API ตอบจำนวนที่คำนวณจากสัญญาณเข้า 5 ออก 2 แล้ว (backend ตรวจการคำนวณใน TestAC_BI_01_HTTP)
    server.use(http.get(OCCUPANCY_URL, () => HttpResponse.json({ count: 3, as_of: FAKE_OCCUPANCY_AS_OF })));

    // When
    renderDashboard();

    // Then
    expect(await screen.findByTestId("occupancy-count")).toHaveTextContent("3");
    expect(screen.getByRole("heading", { name: "จำนวนผู้ใช้บริการในยิมตอนนี้" })).toBeInTheDocument();
  });

  it("AC_BI_04: Given เปิด Dashboard ค้างไว้ When ระบบรับสัญญาณผ่านประตู Then ตัวเลขเปลี่ยนเองโดยไม่ต้องรีเฟรชหน้า", async () => {
    let getRequests = 0;
    server.use(
      http.get(OCCUPANCY_URL, () => {
        getRequests += 1;
        return HttpResponse.json({ count: 3, as_of: FAKE_OCCUPANCY_AS_OF });
      }),
    );

    // Given
    renderDashboard();
    const count = await screen.findByTestId("occupancy-count");
    expect(count).toHaveTextContent("3");
    const source = FakeEventSource.instances.at(-1)!;
    expect(source.url).toBe(OCCUPANCY_STREAM_URL);

    // When: backend ส่ง event จำนวนใหม่ทาง SSE
    act(() => source.emit(OCCUPANCY_EVENT, { count: 4, as_of: "2026-11-12T12:00:05+07:00" }));

    // Then: element เดิม (ไม่ remount) แสดง 4 และไม่มีการโหลดหน้าใหม่ (GET ครั้งเดียว)
    expect(screen.getByTestId("occupancy-count")).toBe(count);
    expect(count).toHaveTextContent("4");
    expect(getRequests).toBe(1);
  });

  it("ไม่ย้อนกลับไปค่าเก่าเมื่อผลของ GET มาถึงหลัง event ที่ใหม่กว่า", async () => {
    let release: () => void = () => {};
    server.use(
      http.get(OCCUPANCY_URL, async () => {
        await new Promise<void>((resolve) => (release = resolve));
        return HttpResponse.json({ count: 3, as_of: "2026-11-12T12:00:00+07:00" });
      }),
    );
    renderDashboard();
    const source = FakeEventSource.instances.at(-1)!;
    act(() => source.emit(OCCUPANCY_EVENT, { count: 5, as_of: "2026-11-12T12:00:10+07:00" }));
    expect(await screen.findByTestId("occupancy-count")).toHaveTextContent("5");

    await act(async () => {
      release();
      await new Promise((resolve) => setTimeout(resolve, 20));
    });

    expect(screen.getByTestId("occupancy-count")).toHaveTextContent("5");
  });

  it("ปิดการเชื่อมต่อ SSE เมื่อออกจากหน้า", async () => {
    const { unmount } = renderDashboard();
    await screen.findByTestId("occupancy-count");
    const source = FakeEventSource.instances.at(-1)!;

    unmount();

    expect(source.closed).toBe(true);
  });

  it("แสดงข้อความเมื่อโหลดจำนวนไม่ได้ (API ตอบ 503)", async () => {
    server.use(occupancyUnavailableHandler);
    renderDashboard();
    expect(await screen.findByRole("alert")).toHaveTextContent("ไม่สามารถโหลดจำนวนผู้ใช้บริการได้");
  });
});
