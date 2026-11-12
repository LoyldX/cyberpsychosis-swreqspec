-- CON-DB-01, FR-BKG-01, FR-BKG-02 (ASM-03, ASM-04): คลาสกลุ่มที่เตรียมไว้ล่วงหน้า มีจำนวนที่นั่งจำกัด
-- trainer_id ยังไม่มี foreign key เพราะตารางเทรนเนอร์เป็นของ 004 ซึ่งยังไม่มี schema (plan D1)
-- ไม่มีฟิลด์สถานะเปิดรับ: คลาสที่ยิมเตรียมไว้คือคลาสที่เปิดรับ (ASM-01)
CREATE TABLE classes (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    trainer_id BIGINT      NOT NULL,
    starts_at  TIMESTAMPTZ NOT NULL,
    ends_at    TIMESTAMPTZ NOT NULL,
    capacity   INTEGER     NOT NULL CHECK (capacity > 0),
    CHECK (ends_at > starts_at)
);

-- ใช้หาคลาสอื่นของเทรนเนอร์คนเดียวกันตอนตรวจเวลาทับซ้อน (DOM-BKG-01, ASM-09)
CREATE INDEX classes_trainer_id_starts_at_idx ON classes (trainer_id, starts_at);

-- FR-BKG-01, NFR-PERF-01 (ASM-05, ASM-06): การจองอ้างสมาชิกใน members ของ 001 ไม่เก็บข้อมูลส่วนตัวซ้ำ
-- UNIQUE (class_id, member_id) กันจองคลาสเดิมซ้ำแม้มีคำขอแข่งกัน (ASM-05)
CREATE TABLE bookings (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    class_id   BIGINT      NOT NULL REFERENCES classes (id),
    member_id  BIGINT      NOT NULL REFERENCES members (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (class_id, member_id)
);
