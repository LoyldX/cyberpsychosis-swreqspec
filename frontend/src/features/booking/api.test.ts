// ตรวจตัวเรียก API การจอง (FR-BKG-01, FR-BKG-02, CON-TECH-02) กับ API จำลอง (msw) ตามสัญญา plan.md ข้อ 4
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { http, HttpResponse } from "msw";
import { clearToken, saveToken } from "../auth/api";
import {
  FAKE_BOOKING_CREATED_AT,
  FAKE_BOOKING_ID,
  FAKE_BUSY_CLASS_ID,
  FAKE_CLASSES,
  FAKE_MEMBER_ID,
  FAKE_REFRESH_TOKEN,
  FAKE_TRAINER_BUSY_MESSAGE,
  classesUnavailableHandler,
  createBookingUnavailableHandler,
} from "../../mocks/handlers";
import { server } from "../../mocks/server";
import { BOOKINGS_URL, createBooking, fetchBooking, fetchClasses } from "./api";

beforeEach(() => saveToken(FAKE_REFRESH_TOKEN));
afterEach(() => clearToken());

describe("fetchClasses", () => {
  it("ส่ง token ของ 001 และอ่านรายการคลาส", async () => {
    await expect(fetchClasses()).resolves.toEqual({ kind: "ok", classes: FAKE_CLASSES });
  });

  it("ไม่มี token ได้ unauthorized", async () => {
    clearToken();
    await expect(fetchClasses()).resolves.toEqual({ kind: "unauthorized" });
  });

  it("503 ได้ unavailable (ASM-07)", async () => {
    server.use(classesUnavailableHandler);
    await expect(fetchClasses()).resolves.toEqual({ kind: "unavailable" });
  });
});

describe("createBooking", () => {
  it("201 คืนการจองที่ได้ (FR-BKG-01)", async () => {
    await expect(createBooking(FAKE_CLASSES[0].id)).resolves.toEqual({
      kind: "ok",
      booking: {
        id: FAKE_BOOKING_ID,
        class_id: FAKE_CLASSES[0].id,
        member_id: FAKE_MEMBER_ID,
        created_at: FAKE_BOOKING_CREATED_AT,
      },
    });
  });

  it("409 แปลงรหัสเหตุผลและข้อความจาก backend (DOM-BKG-01, ASM-05)", async () => {
    await expect(createBooking(FAKE_BUSY_CLASS_ID)).resolves.toEqual({
      kind: "rejected",
      reason: "trainer_busy",
      message: FAKE_TRAINER_BUSY_MESSAGE,
    });
  });

  it("404 class_not_found เป็น rejected", async () => {
    const result = await createBooking(999);
    expect(result).toMatchObject({ kind: "rejected", reason: "class_not_found" });
  });

  it("รหัสที่ไม่รู้จักหรือ 503 ได้ unavailable (ASM-07)", async () => {
    server.use(createBookingUnavailableHandler);
    await expect(createBooking(FAKE_CLASSES[0].id)).resolves.toEqual({ kind: "unavailable" });

    server.use(http.post(BOOKINGS_URL, () => HttpResponse.json({ error: "weird" }, { status: 409 })));
    await expect(createBooking(FAKE_CLASSES[0].id)).resolves.toEqual({ kind: "unavailable" });
  });

  it("เครือข่ายล้มได้ unavailable", async () => {
    server.use(http.post(BOOKINGS_URL, () => HttpResponse.error()));
    await expect(createBooking(FAKE_CLASSES[0].id)).resolves.toEqual({ kind: "unavailable" });
  });

  it("ไม่มี token ได้ unauthorized", async () => {
    clearToken();
    await expect(createBooking(FAKE_CLASSES[0].id)).resolves.toEqual({ kind: "unauthorized" });
  });
});

describe("fetchBooking", () => {
  it("อ่านการจองของตนเองทีละรายการ (AC-BKG-01 เรียกดูได้)", async () => {
    await createBooking(FAKE_CLASSES[0].id);
    await expect(fetchBooking(FAKE_BOOKING_ID)).resolves.toMatchObject({
      kind: "ok",
      booking: { id: FAKE_BOOKING_ID, class_id: FAKE_CLASSES[0].id },
    });
  });

  it("404 ได้ not_found (plan D5)", async () => {
    await expect(fetchBooking(FAKE_BOOKING_ID + 1)).resolves.toEqual({ kind: "not_found" });
  });
});
