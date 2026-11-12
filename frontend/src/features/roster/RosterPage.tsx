import { useEffect, useState, type FormEvent } from "react";
import { cancelEntry, createEntry, deleteEntry, fetchRoster, fetchTrainers, updateEntry } from "./api";
import type { DeleteResult, EntryResult, RosterEntry, RosterKind, Trainer } from "./types";

export const WEEKDAYS = ["จันทร์", "อังคาร", "พุธ", "พฤหัสบดี", "ศุกร์", "เสาร์", "อาทิตย์"]; // ISO 1..7 (ASM-03)
const KIND_LABEL: Record<RosterKind, string> = { work: "เวลาเข้างาน", teach: "เวลาสอน" };
export const UNAVAILABLE_MESSAGE = "ระบบไม่พร้อมให้บริการ กรุณาลองใหม่อีกครั้ง";

type SlotForm = { weekday: string; start_time: string; end_time: string };
const emptySlot: SlotForm = { weekday: "1", start_time: "", end_time: "" };

function describe(entry: RosterEntry): string {
  return `${WEEKDAYS[entry.weekday - 1]} ${entry.start_time}–${entry.end_time} ${KIND_LABEL[entry.kind]}`;
}

function toSlot(form: SlotForm) {
  return { weekday: Number(form.weekday), start_time: form.start_time, end_time: form.end_time };
}

// ช่องวันและเวลาที่ใช้ทั้งฟอร์มเพิ่มและฟอร์มแก้ไข
function SlotFields({ value, onChange }: { value: SlotForm; onChange: (next: SlotForm) => void }) {
  return (
    <>
      <label>
        วัน
        <select value={value.weekday} onChange={(e) => onChange({ ...value, weekday: e.target.value })}>
          {WEEKDAYS.map((name, i) => (
            <option key={name} value={String(i + 1)}>
              {name}
            </option>
          ))}
        </select>
      </label>
      <label>
        เวลาเริ่ม
        <input type="time" value={value.start_time} onChange={(e) => onChange({ ...value, start_time: e.target.value })} />
      </label>
      <label>
        เวลาจบ
        <input type="time" value={value.end_time} onChange={(e) => onChange({ ...value, end_time: e.target.value })} />
      </label>
    </>
  );
}

// รองรับ FR-ROST-01, CON-TECH-02 (ASM-01, ASM-05, ASM-06, ASM-08, plan D7): หน้าจัดเวลาเข้างานและเวลาสอนของเทรนเนอร์
// เปิดได้โดยไม่ต้องล็อกอิน; เพิ่ม/แก้/ยกเลิก/ลบทีละรายการ; บันทึกไม่ได้แสดงเหตุผล; ระบบไม่พร้อมแจ้งให้ลองใหม่และคงค่าในฟอร์ม
export function RosterPage() {
  const [trainers, setTrainers] = useState<Trainer[]>([]);
  const [trainerId, setTrainerId] = useState<number | null>(null);
  const [entries, setEntries] = useState<RosterEntry[]>([]);
  const [kind, setKind] = useState<RosterKind>("work");
  const [newSlot, setNewSlot] = useState<SlotForm>(emptySlot);
  const [editing, setEditing] = useState<{ id: number; slot: SlotForm } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    let active = true;
    fetchTrainers().then((result) => {
      if (!active) return;
      if (result.kind === "ok") setTrainers(result.trainers);
      else setError(UNAVAILABLE_MESSAGE);
    });
    return () => {
      active = false;
    };
  }, []);

  async function loadRoster(id: number) {
    const result = await fetchRoster(id);
    if (result.kind === "ok") setEntries(result.entries);
    else setError(result.kind === "rejected" ? result.message : UNAVAILABLE_MESSAGE);
  }

  async function selectTrainer(value: string) {
    setError(null);
    setEditing(null);
    const id = value ? Number(value) : null;
    setTrainerId(id);
    setEntries([]);
    if (id !== null) await loadRoster(id);
  }

  // ทำคำสั่งหนึ่งครั้ง: สำเร็จโหลดตารางใหม่และคืน true, ไม่สำเร็จแสดงเหตุผลและคืน false (ค่าในฟอร์มคงเดิม ASM-08)
  async function run(action: () => Promise<EntryResult | DeleteResult>): Promise<boolean> {
    if (trainerId === null) return false;
    setBusy(true);
    setError(null);
    const result = await action();
    setBusy(false);
    if (result.kind === "ok") {
      await loadRoster(trainerId);
      return true;
    }
    setError(result.kind === "rejected" ? result.message : UNAVAILABLE_MESSAGE);
    return false;
  }

  async function submitNew(event: FormEvent) {
    event.preventDefault();
    if (trainerId === null) return;
    await run(() => createEntry(trainerId, { kind, ...toSlot(newSlot) }));
  }

  async function submitEdit(event: FormEvent) {
    event.preventDefault();
    if (!editing) return;
    if (await run(() => updateEntry(editing.id, toSlot(editing.slot)))) setEditing(null);
  }

  return (
    <main>
      <h1>จัดตารางงานเทรนเนอร์</h1>
      <label>
        เทรนเนอร์
        <select value={trainerId ?? ""} onChange={(e) => selectTrainer(e.target.value)}>
          <option value="">เลือกเทรนเนอร์</option>
          {trainers.map((t) => (
            <option key={t.id} value={String(t.id)}>
              {t.name}
            </option>
          ))}
        </select>
      </label>

      {error && <p role="alert">{error}</p>}

      {trainerId !== null && (
        <>
          {entries.length === 0 ? (
            <p>ยังไม่มีตารางงาน</p>
          ) : (
            <ul aria-label="ตารางงาน">
              {entries.map((entry) => (
                <li key={entry.id}>
                  <span>
                    {describe(entry)}
                    {entry.cancelled && " (ยกเลิกแล้ว)"}
                  </span>{" "}
                  <button
                    type="button"
                    disabled={busy}
                    onClick={() =>
                      setEditing({
                        id: entry.id,
                        slot: { weekday: String(entry.weekday), start_time: entry.start_time, end_time: entry.end_time },
                      })
                    }
                  >
                    แก้ไข
                  </button>{" "}
                  {!entry.cancelled && (
                    <button type="button" disabled={busy} onClick={() => run(() => cancelEntry(entry.id))}>
                      ยกเลิก
                    </button>
                  )}{" "}
                  <button type="button" disabled={busy} onClick={() => run(() => deleteEntry(entry.id))}>
                    ลบ
                  </button>
                </li>
              ))}
            </ul>
          )}

          {editing && (
            <form aria-label="แก้ไขรายการ" onSubmit={submitEdit}>
              <SlotFields value={editing.slot} onChange={(slot) => setEditing({ ...editing, slot })} />
              <button type="submit" disabled={busy}>
                บันทึกการแก้ไข
              </button>
              <button type="button" onClick={() => setEditing(null)}>
                ไม่แก้ไข
              </button>
            </form>
          )}

          <form aria-label="เพิ่มรายการ" onSubmit={submitNew}>
            <label>
              ประเภท
              <select value={kind} onChange={(e) => setKind(e.target.value as RosterKind)}>
                <option value="work">{KIND_LABEL.work}</option>
                <option value="teach">{KIND_LABEL.teach}</option>
              </select>
            </label>
            <SlotFields value={newSlot} onChange={setNewSlot} />
            <button type="submit" disabled={busy}>
              เพิ่ม
            </button>
          </form>
        </>
      )}
    </main>
  );
}
