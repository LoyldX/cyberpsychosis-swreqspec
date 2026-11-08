-- CON-DB-01, FR-AUTH-01: ตาราง members เก็บสมาชิกใน PostgreSQL (plan.md §3)
-- ไม่เก็บรหัสผ่านจริง เก็บเฉพาะ password_hash (ผลจาก bcrypt) และไม่มีฟิลด์ชื่อ/อีเมล/เบอร์ เพราะ spec 001 ไม่ใช้
CREATE TABLE members (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    student_id    TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
