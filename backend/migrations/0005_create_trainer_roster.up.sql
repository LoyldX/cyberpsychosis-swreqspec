-- CON-DB-01, FR-ROST-01 (ASM-09): ข้อมูลเทรนเนอร์เตรียมไว้ล่วงหน้า เก็บแค่รหัสและชื่อ
CREATE TABLE trainers (
    id   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL
);

-- CON-DB-01, FR-ROST-01 (ASM-01, ASM-03, ASM-04): ตารางงานซ้ำทุกสัปดาห์ของเทรนเนอร์แต่ละคน
-- kind: work = เวลาเข้างาน, teach = เวลาสอน (แยกกัน ASM-04)
-- weekday: ISO จันทร์ = 1 ... อาทิตย์ = 7; start/end เป็นเวลา Asia/Bangkok ในวันเดียว (end_time > start_time, ไม่ข้ามเที่ยงคืน ASM-03)
-- cancelled: ยกเลิก = ยังเก็บไว้แต่ไม่นับ ส่วนลบ = ลบแถว (ASM-01)
CREATE TABLE roster_entries (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    trainer_id BIGINT   NOT NULL REFERENCES trainers (id),
    kind       TEXT     NOT NULL CHECK (kind IN ('work', 'teach')),
    weekday    SMALLINT NOT NULL CHECK (weekday BETWEEN 1 AND 7),
    start_time TIME     NOT NULL,
    end_time   TIME     NOT NULL,
    cancelled  BOOLEAN  NOT NULL DEFAULT false,
    CHECK (end_time > start_time)
);

-- ใช้ดึงรายการของเทรนเนอร์ในวันหนึ่งตอนตรวจกฎและ Covers (ASM-02, ASM-10)
CREATE INDEX roster_entries_trainer_id_weekday_idx ON roster_entries (trainer_id, weekday);
