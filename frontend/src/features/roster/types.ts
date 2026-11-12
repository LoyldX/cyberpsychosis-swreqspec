// รองรับ FR-ROST-01, CON-TECH-02: ชนิดข้อมูลตามสัญญา API ใน specs/004-trainer-roster/plan.md ข้อ 4

// เทรนเนอร์ที่เตรียมไว้ล่วงหน้า (ASM-09)
export type Trainer = { id: number; name: string };

// work = เวลาเข้างาน, teach = เวลาสอน (แยกกัน ASM-04)
export type RosterKind = "work" | "teach";

// รายการซ้ำทุกสัปดาห์: weekday ISO (จันทร์ = 1 … อาทิตย์ = 7), เวลา "HH:MM" ในวันเดียว (ASM-03); cancelled = ยกเลิกแต่ยังเก็บไว้ (ASM-01)
export type RosterEntry = {
  id: number;
  trainer_id: number;
  kind: RosterKind;
  weekday: number;
  start_time: string;
  end_time: string;
  cancelled: boolean;
};

export type RosterSlot = { weekday: number; start_time: string; end_time: string };

// รหัสเหตุผลที่ backend ตอบเมื่อบันทึกไม่ได้ (ASM-02, ASM-04, ASM-07)
export type RosterRejection =
  | "invalid_request"
  | "trainer_not_found"
  | "entry_not_found"
  | "overlap"
  | "teach_outside_work"
  | "booked_class_outside_work";

export type Rejected = { kind: "rejected"; reason: RosterRejection; message: string }; // HTTP 400/404/409
export type Unavailable = { kind: "unavailable" }; // HTTP 503, สถานะอื่น หรือเครือข่ายล้ม (ASM-08)

export type TrainersResult = { kind: "ok"; trainers: Trainer[] } | Unavailable;
export type RosterResult = { kind: "ok"; entries: RosterEntry[] } | Rejected | Unavailable;
export type EntryResult = { kind: "ok"; entry: RosterEntry } | Rejected | Unavailable;
export type DeleteResult = { kind: "ok" } | Rejected | Unavailable;
