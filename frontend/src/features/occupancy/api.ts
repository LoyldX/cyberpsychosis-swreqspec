// รองรับ FR-BI-01, NFR-PERF-02, CON-TECH-02: เรียก API ของ Occupancy Dashboard (plan §4.1)
import type { OccupancySnapshot } from "./types";

export const OCCUPANCY_URL = "/api/occupancy";
export const OCCUPANCY_STREAM_URL = "/api/occupancy/stream";
/** ชื่อ event ของ SSE ที่ backend ส่ง (plan §4.1) */
export const OCCUPANCY_EVENT = "occupancy";

/** ส่วนของ EventSource ที่ใช้ แยกออกมาเพื่อให้ test ส่งตัวจำลองเข้ามาได้ (jsdom ไม่มี EventSource) */
export type EventSourceLike = {
  addEventListener(type: string, listener: (event: MessageEvent) => void): void;
  close(): void;
};

export type CreateEventSource = (url: string) => EventSourceLike;

const createBrowserEventSource: CreateEventSource = (url) =>
  new EventSource(url) as unknown as EventSourceLike;

function isSnapshot(value: unknown): value is OccupancySnapshot {
  const v = value as Partial<OccupancySnapshot> | null;
  return typeof v?.count === "number" && typeof v?.as_of === "string";
}

// รองรับ FR-BI-01: อ่านจำนวนผู้ใช้บริการปัจจุบันตอนเปิดหน้า Dashboard (AC-BI-01)
export async function fetchOccupancy(): Promise<OccupancySnapshot> {
  const response = await fetch(OCCUPANCY_URL);
  if (!response.ok) {
    throw new Error("ไม่สามารถโหลดจำนวนผู้ใช้บริการได้");
  }
  const body: unknown = await response.json();
  if (!isSnapshot(body)) {
    throw new Error("ข้อมูลจำนวนผู้ใช้บริการไม่ถูกต้อง");
  }
  return body;
}

// รองรับ NFR-PERF-02 (ASM-06): ฟังจำนวนใหม่ทาง SSE ทุกครั้งที่มีการผ่านประตู คืนฟังก์ชันปิดการเชื่อมต่อ (AC-BI-04)
export function subscribeOccupancy(
  onSnapshot: (snapshot: OccupancySnapshot) => void,
  createSource: CreateEventSource = createBrowserEventSource,
): () => void {
  const source = createSource(OCCUPANCY_STREAM_URL);
  source.addEventListener(OCCUPANCY_EVENT, (event) => {
    let data: unknown;
    try {
      data = JSON.parse(String(event.data));
    } catch {
      return;
    }
    if (isSnapshot(data)) {
      onSnapshot(data);
    }
  });
  return () => source.close();
}
