import { useEffect, useState } from "react";
import { fetchOccupancy, subscribeOccupancy, type CreateEventSource } from "./api";
import type { OccupancySnapshot } from "./types";

type Props = {
  /** ใช้ใน test เพื่อส่ง EventSource จำลองเข้ามา (ค่าเริ่มต้นคือของเบราว์เซอร์) */
  createSource?: CreateEventSource;
};

/** เก็บค่าที่ใหม่กว่าตาม as_of — กันกรณีผลของ GET มาถึงหลัง event ของ SSE ที่ใหม่กว่า */
function newer(prev: OccupancySnapshot | null, next: OccupancySnapshot): OccupancySnapshot {
  if (!prev) return next;
  return Date.parse(next.as_of) >= Date.parse(prev.as_of) ? next : prev;
}

function formatTime(asOf: string): string {
  return new Date(asOf).toLocaleTimeString("th-TH", { timeZone: "Asia/Bangkok" });
}

// รองรับ FR-BI-01, NFR-PERF-02, CON-TECH-02 (ASM-07): หน้า Dashboard แสดงจำนวนผู้ใช้บริการปัจจุบัน
// โหลดจำนวนตอนเปิดหน้า (AC-BI-01) แล้วรับจำนวนใหม่ทาง SSE โดยไม่ต้องรีเฟรชหน้า (AC-BI-04) เปิดได้โดยไม่ต้องล็อกอิน
export function OccupancyDashboardPage({ createSource }: Props) {
  const [snapshot, setSnapshot] = useState<OccupancySnapshot | null>(null);
  const [loadFailed, setLoadFailed] = useState(false);

  useEffect(() => {
    let active = true;
    const accept = (next: OccupancySnapshot) => {
      if (!active) return;
      setSnapshot((prev) => newer(prev, next));
      setLoadFailed(false);
    };
    const unsubscribe = subscribeOccupancy(accept, createSource);
    fetchOccupancy()
      .then(accept)
      .catch(() => {
        if (active) setLoadFailed(true);
      });
    return () => {
      active = false;
      unsubscribe();
    };
  }, [createSource]);

  return (
    <main>
      <h1>จำนวนผู้ใช้บริการในยิมตอนนี้</h1>
      {snapshot ? (
        <>
          <p role="status" aria-live="polite" data-testid="occupancy-count" style={{ fontSize: "4rem", margin: 0 }}>
            {snapshot.count}
          </p>
          <p>อัปเดตล่าสุด {formatTime(snapshot.as_of)}</p>
        </>
      ) : loadFailed ? (
        <p role="alert">ไม่สามารถโหลดจำนวนผู้ใช้บริการได้ กรุณาลองใหม่ภายหลัง</p>
      ) : (
        <p>กำลังโหลด...</p>
      )}
    </main>
  );
}
