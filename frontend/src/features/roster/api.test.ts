// ตรวจตัวเรียก API ตารางงาน (FR-ROST-01, CON-TECH-02) กับ API จำลอง (msw) ตามสัญญา plan.md ข้อ 4
import { beforeEach, describe, expect, it } from "vitest";
import { http, HttpResponse } from "msw";
import {
  FAKE_TRAINERS,
  createRosterUnavailableHandler,
  resetFakeRoster,
  trainersUnavailableHandler,
} from "../../mocks/handlers";
import { server } from "../../mocks/server";
import {
  TRAINERS_URL,
  cancelEntry,
  createEntry,
  deleteEntry,
  fetchRoster,
  fetchTrainers,
  updateEntry,
} from "./api";

beforeEach(() => resetFakeRoster());

describe("roster API", () => {
  it("อ่านรายชื่อเทรนเนอร์โดยไม่ต้องล็อกอิน (ASM-06, ASM-09)", async () => {
    await expect(fetchTrainers()).resolves.toEqual({ kind: "ok", trainers: FAKE_TRAINERS });
  });

  it("เพิ่มรายการแล้วเรียกดูได้ (AC-ROST-01)", async () => {
    const created = await createEntry(2, { kind: "work", weekday: 3, start_time: "13:00", end_time: "17:00" });
    expect(created).toMatchObject({ kind: "ok", entry: { trainer_id: 2, kind: "work", weekday: 3, cancelled: false } });
    const roster = await fetchRoster(2);
    expect(roster).toMatchObject({ kind: "ok", entries: [{ start_time: "13:00", end_time: "17:00" }] });
  });

  it("แก้ ยกเลิก และลบรายการ (AC-ROST-03, ASM-01)", async () => {
    await expect(updateEntry(1, { weekday: 1, start_time: "08:00", end_time: "12:00" })).resolves.toMatchObject({
      kind: "ok",
      entry: { id: 1, start_time: "08:00" },
    });
    await expect(cancelEntry(1)).resolves.toMatchObject({ kind: "ok", entry: { id: 1, cancelled: true } });
    await expect(deleteEntry(1)).resolves.toEqual({ kind: "ok" });
    await expect(fetchRoster(1)).resolves.toEqual({ kind: "ok", entries: [] });
  });

  it("409 แปลงรหัสเหตุผลและข้อความจาก backend (ASM-02, ASM-04)", async () => {
    await expect(
      createEntry(1, { kind: "work", weekday: 1, start_time: "11:00", end_time: "13:00" }),
    ).resolves.toMatchObject({ kind: "rejected", reason: "overlap" });
    await expect(
      createEntry(1, { kind: "teach", weekday: 2, start_time: "10:00", end_time: "11:00" }),
    ).resolves.toMatchObject({ kind: "rejected", reason: "teach_outside_work" });
  });

  it("400 และ 404 เป็น rejected", async () => {
    await expect(
      createEntry(1, { kind: "work", weekday: 1, start_time: "12:00", end_time: "09:00" }),
    ).resolves.toMatchObject({ kind: "rejected", reason: "invalid_request" });
    await expect(fetchRoster(99)).resolves.toMatchObject({ kind: "rejected", reason: "trainer_not_found" });
    await expect(deleteEntry(999)).resolves.toMatchObject({ kind: "rejected", reason: "entry_not_found" });
  });

  it("503 รหัสที่ไม่รู้จัก หรือเครือข่ายล้ม ได้ unavailable (ASM-08)", async () => {
    server.use(trainersUnavailableHandler, createRosterUnavailableHandler);
    await expect(fetchTrainers()).resolves.toEqual({ kind: "unavailable" });
    await expect(
      createEntry(1, { kind: "work", weekday: 2, start_time: "09:00", end_time: "12:00" }),
    ).resolves.toEqual({ kind: "unavailable" });

    server.use(http.get(`${TRAINERS_URL}/:id/roster`, () => HttpResponse.json({ error: "weird" }, { status: 409 })));
    await expect(fetchRoster(1)).resolves.toEqual({ kind: "unavailable" });
    server.use(http.get(`${TRAINERS_URL}/:id/roster`, () => HttpResponse.error()));
    await expect(fetchRoster(1)).resolves.toEqual({ kind: "unavailable" });
  });
});
