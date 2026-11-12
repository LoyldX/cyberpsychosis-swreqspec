import { useEffect, useState } from "react";
import { createBooking, fetchBooking, fetchClasses } from "./api";
import type { Booking, BookingRejection, ClassSummary } from "./types";

// ข้อความสำรองเมื่อ backend ไม่ได้ส่ง message มา (ASM-05)
const FALLBACK_REJECTION: Record<BookingRejection, string> = {
  invalid_request: "ข้อมูลการจองไม่ถูกต้อง",
  class_not_found: "ไม่พบคลาสนี้",
  class_started: "คลาสนี้เริ่มแล้ว จองไม่ได้",
  already_booked: "คุณจองคลาสนี้ไว้แล้ว",
  class_full: "คลาสนี้เต็มแล้ว",
  trainer_off_shift: "จองไม่สำเร็จ เพราะคลาสอยู่นอกเวลาปฏิบัติงานของเทรนเนอร์",
  trainer_busy: "จองไม่สำเร็จ เพราะเทรนเนอร์ไม่ว่าง",
};
export const UNAVAILABLE_MESSAGE = "ระบบไม่พร้อมให้บริการ กรุณาลองใหม่อีกครั้ง";
export const UNAUTHORIZED_MESSAGE = "กรุณาเข้าสู่ระบบใหม่อีกครั้ง";

const timeFormat = new Intl.DateTimeFormat("th-TH", {
  timeZone: "Asia/Bangkok",
  dateStyle: "medium",
  timeStyle: "short",
});

function formatRange(cls: ClassSummary): string {
  return `${timeFormat.format(new Date(cls.starts_at))} – ${timeFormat.format(new Date(cls.ends_at))}`;
}

type LoadState =
  | { kind: "loading" }
  | { kind: "ok"; classes: ClassSummary[] }
  | { kind: "error"; message: string };

// รองรับ FR-BKG-01, FR-BKG-02, CON-TECH-02 (ASM-05, ASM-07, plan D7): หน้าให้สมาชิกที่ล็อกอินแล้วเลือกและจองคลาส
// สำเร็จแล้วดึงการจองนั้นมาแสดง (AC-BKG-01 "เรียกดูการจองนั้นได้") ไม่สำเร็จแสดงเหตุผลจาก backend
export function ClassBookingPage() {
  const [load, setLoad] = useState<LoadState>({ kind: "loading" });
  const [submitting, setSubmitting] = useState(false);
  const [booked, setBooked] = useState<{ booking: Booking; cls: ClassSummary | undefined } | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    fetchClasses().then((result) => {
      if (!active) return;
      if (result.kind === "ok") setLoad({ kind: "ok", classes: result.classes });
      else setLoad({ kind: "error", message: result.kind === "unauthorized" ? UNAUTHORIZED_MESSAGE : UNAVAILABLE_MESSAGE });
    });
    return () => {
      active = false;
    };
  }, []);

  const classes = load.kind === "ok" ? load.classes : [];

  async function book(cls: ClassSummary) {
    setSubmitting(true);
    setError(null);
    setBooked(null);
    const result = await createBooking(cls.id);
    if (result.kind === "ok") {
      // เรียกดูการจองที่เพิ่งได้จาก backend (AC-BKG-01) ถ้าดึงไม่ได้ยังแสดงผลจากการจองเอง
      const fetched = await fetchBooking(result.booking.id);
      const booking = fetched.kind === "ok" ? fetched.booking : result.booking;
      setBooked({ booking, cls: classes.find((c) => c.id === booking.class_id) });
    } else if (result.kind === "rejected") {
      setError(result.message || FALLBACK_REJECTION[result.reason]);
    } else {
      setError(result.kind === "unauthorized" ? UNAUTHORIZED_MESSAGE : UNAVAILABLE_MESSAGE);
    }
    setSubmitting(false);
  }

  return (
    <main>
      <h1>จองคลาสออกกำลังกาย</h1>

      {booked && (
        <section role="status" aria-label="ผลการจอง">
          <h2>จองสำเร็จ</h2>
          <p>หมายเลขการจอง {booked.booking.id}</p>
          {booked.cls && (
            <p>
              คลาสของเทรนเนอร์ #{booked.cls.trainer_id} เวลา {formatRange(booked.cls)}
            </p>
          )}
        </section>
      )}
      {error && <p role="alert">{error}</p>}

      {load.kind === "loading" && <p>กำลังโหลดรายการคลาส...</p>}
      {load.kind === "error" && <p role="alert">{load.message}</p>}
      {load.kind === "ok" && classes.length === 0 && <p>ยังไม่มีคลาสที่เปิดให้จอง</p>}
      {classes.length > 0 && (
        <ul aria-label="คลาสที่เปิดให้จอง">
          {classes.map((cls) => (
            <li key={cls.id}>
              <span>
                เทรนเนอร์ #{cls.trainer_id} · {formatRange(cls)} · รับ {cls.capacity} คน
              </span>{" "}
              <button
                type="button"
                onClick={() => book(cls)}
                disabled={submitting}
                aria-label={`จองคลาส ${formatRange(cls)}`}
              >
                จอง
              </button>
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}
