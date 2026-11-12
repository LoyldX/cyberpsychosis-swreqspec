// ตรวจตัวเรียก API ของ Occupancy Dashboard (FR-BI-01, NFR-PERF-02) ด้วย API จำลอง (msw) และ EventSource จำลอง
import { describe, expect, it, vi } from "vitest";
import { FAKE_OCCUPANCY_AS_OF, FAKE_OCCUPANCY_COUNT, occupancyUnavailableHandler } from "../../mocks/handlers";
import { server } from "../../mocks/server";
import {
  OCCUPANCY_EVENT,
  OCCUPANCY_STREAM_URL,
  fetchOccupancy,
  subscribeOccupancy,
  type EventSourceLike,
} from "./api";

/** EventSource จำลอง: test เรียก emit เพื่อส่ง event เหมือน backend */
class FakeEventSource implements EventSourceLike {
  listeners = new Map<string, ((event: MessageEvent) => void)[]>();
  closed = false;
  constructor(public url: string) {}
  addEventListener(type: string, listener: (event: MessageEvent) => void) {
    this.listeners.set(type, [...(this.listeners.get(type) ?? []), listener]);
  }
  close() {
    this.closed = true;
  }
  emit(type: string, data: string) {
    for (const listener of this.listeners.get(type) ?? []) {
      listener(new MessageEvent(type, { data }));
    }
  }
}

describe("fetchOccupancy", () => {
  it("อ่าน {count, as_of} จาก GET /api/occupancy", async () => {
    await expect(fetchOccupancy()).resolves.toEqual({
      count: FAKE_OCCUPANCY_COUNT,
      as_of: FAKE_OCCUPANCY_AS_OF,
    });
  });

  it("โยน error เมื่อ API ตอบ 503", async () => {
    server.use(occupancyUnavailableHandler);
    await expect(fetchOccupancy()).rejects.toThrow();
  });
});

describe("subscribeOccupancy", () => {
  it("เปิด stream ที่ /api/occupancy/stream และเรียก callback เมื่อได้ event occupancy", () => {
    let source: FakeEventSource | undefined;
    const onSnapshot = vi.fn();
    subscribeOccupancy(onSnapshot, (url) => (source = new FakeEventSource(url)));

    expect(source?.url).toBe(OCCUPANCY_STREAM_URL);
    source!.emit(OCCUPANCY_EVENT, JSON.stringify({ count: 4, as_of: FAKE_OCCUPANCY_AS_OF }));

    expect(onSnapshot).toHaveBeenCalledWith({ count: 4, as_of: FAKE_OCCUPANCY_AS_OF });
  });

  it("ไม่เรียก callback เมื่อข้อมูลไม่ถูกต้อง หรือเป็น event ชื่ออื่น", () => {
    let source: FakeEventSource | undefined;
    const onSnapshot = vi.fn();
    subscribeOccupancy(onSnapshot, (url) => (source = new FakeEventSource(url)));

    source!.emit(OCCUPANCY_EVENT, "not json");
    source!.emit(OCCUPANCY_EVENT, JSON.stringify({ count: "4" }));
    source!.emit("message", JSON.stringify({ count: 4, as_of: FAKE_OCCUPANCY_AS_OF }));

    expect(onSnapshot).not.toHaveBeenCalled();
  });

  it("ฟังก์ชันที่คืนมาปิดการเชื่อมต่อ", () => {
    let source: FakeEventSource | undefined;
    const unsubscribe = subscribeOccupancy(vi.fn(), (url) => (source = new FakeEventSource(url)));

    unsubscribe();

    expect(source?.closed).toBe(true);
  });
});
