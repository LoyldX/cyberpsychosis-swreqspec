// รองรับ FR-ROST-01, CON-TECH-02: เรียก API ตารางงานตามสัญญา plan.md ข้อ 4 — ไม่ต้องล็อกอิน (ASM-06)
import type {
  DeleteResult,
  EntryResult,
  Rejected,
  RosterEntry,
  RosterKind,
  RosterRejection,
  RosterResult,
  RosterSlot,
  Trainer,
  TrainersResult,
  Unavailable,
} from "./types";

export const TRAINERS_URL = "/api/trainers";
export const ROSTER_ENTRY_URL = "/api/roster";

const REJECTIONS: readonly RosterRejection[] = [
  "invalid_request",
  "trainer_not_found",
  "entry_not_found",
  "overlap",
  "teach_outside_work",
  "booked_class_outside_work",
];
const UNAVAILABLE: Unavailable = { kind: "unavailable" };

// ส่งคำขอ ถ้าเครือข่ายล้มคืน null (ถือว่า unavailable ตาม ASM-08)
async function send(url: string, init: RequestInit = {}): Promise<Response | null> {
  try {
    return await fetch(url, init);
  } catch {
    return null;
  }
}

// แปลง 400/404/409 ที่มีรหัสเหตุผลที่รู้จักเป็น rejected ไม่เช่นนั้นเป็น unavailable
async function rejection(response: Response | null): Promise<Rejected | Unavailable> {
  if (!response || ![400, 404, 409].includes(response.status)) return UNAVAILABLE;
  const body = (await response.json().catch(() => null)) as { error?: string; message?: string } | null;
  const reason = REJECTIONS.find((r) => r === body?.error);
  return reason ? { kind: "rejected", reason, message: body?.message ?? "" } : UNAVAILABLE;
}

function jsonInit(method: string, body: unknown): RequestInit {
  return { method, headers: { "Content-Type": "application/json" }, body: JSON.stringify(body) };
}

// รองรับ FR-ROST-01 (ASM-09): GET /api/trainers
export async function fetchTrainers(): Promise<TrainersResult> {
  const response = await send(TRAINERS_URL);
  if (response?.status === 200) return { kind: "ok", trainers: (await response.json()) as Trainer[] };
  return UNAVAILABLE;
}

// รองรับ FR-ROST-01 (AC-ROST-01..04 เรียกดูได้): GET /api/trainers/:id/roster
export async function fetchRoster(trainerId: number): Promise<RosterResult> {
  const response = await send(`${TRAINERS_URL}/${trainerId}/roster`);
  if (response?.status === 200) return { kind: "ok", entries: (await response.json()) as RosterEntry[] };
  return rejection(response);
}

// รองรับ FR-ROST-01 (AC-ROST-01, 02; ASM-02, ASM-04): POST /api/trainers/:id/roster
export async function createEntry(
  trainerId: number,
  input: RosterSlot & { kind: RosterKind },
): Promise<EntryResult> {
  const response = await send(`${TRAINERS_URL}/${trainerId}/roster`, jsonInit("POST", input));
  if (response?.status === 201) return { kind: "ok", entry: (await response.json()) as RosterEntry };
  return rejection(response);
}

// รองรับ FR-ROST-01 (AC-ROST-03; ASM-05, ASM-07): PATCH /api/roster/:entryId
export async function updateEntry(entryId: number, slot: RosterSlot): Promise<EntryResult> {
  const response = await send(`${ROSTER_ENTRY_URL}/${entryId}`, jsonInit("PATCH", slot));
  if (response?.status === 200) return { kind: "ok", entry: (await response.json()) as RosterEntry };
  return rejection(response);
}

// รองรับ FR-ROST-01 (ASM-01, ASM-07): POST /api/roster/:entryId/cancel
export async function cancelEntry(entryId: number): Promise<EntryResult> {
  const response = await send(`${ROSTER_ENTRY_URL}/${entryId}/cancel`, { method: "POST" });
  if (response?.status === 200) return { kind: "ok", entry: (await response.json()) as RosterEntry };
  return rejection(response);
}

// รองรับ FR-ROST-01 (ASM-01, ASM-07): DELETE /api/roster/:entryId
export async function deleteEntry(entryId: number): Promise<DeleteResult> {
  const response = await send(`${ROSTER_ENTRY_URL}/${entryId}`, { method: "DELETE" });
  if (response?.status === 204) return { kind: "ok" };
  return rejection(response);
}
