-- ข้อมูลคลาสตัวอย่างสำหรับ dev เท่านั้น (005 ASM-04: เตรียมไว้ล่วงหน้า, plan D9) — ไม่ใช่ migration
-- รันด้วยมือหลัง backend/seeds/004_trainers.sql: psql "$DATABASE_URL" -f backend/seeds/005_classes.sql  (รันซ้ำได้)
-- ทุกคลาสเป็นของเทรนเนอร์ 1001 (ครูเอ ตัวอย่าง) ในวันจันทร์ถัดไปตามเวลา Asia/Bangkok เพื่อให้ยังไม่ถึงเวลาเริ่ม (ASM-05)
--   10:00–11:00 และ 10:30–11:30 ทับซ้อนกัน (ใช้ตรวจ AC-BKG-02), 11:30–12:30 อยู่นอกกะถ้าเวลาเข้างานคือ 09:00–12:00 (AC-BKG-04)
-- หมายเหตุ: test ของ backend ล้างตาราง classes ในฐานข้อมูล dev (plan D8) ต้องรันไฟล์นี้ใหม่หลังรัน test
WITH next_monday AS (
    SELECT (date_trunc('week', now() AT TIME ZONE 'Asia/Bangkok') + interval '7 days') AS day
), sample (start_at, end_at, capacity) AS (
    SELECT (day + time '10:00') AT TIME ZONE 'Asia/Bangkok', (day + time '11:00') AT TIME ZONE 'Asia/Bangkok', 10 FROM next_monday
    UNION ALL
    SELECT (day + time '10:30') AT TIME ZONE 'Asia/Bangkok', (day + time '11:30') AT TIME ZONE 'Asia/Bangkok', 8 FROM next_monday
    UNION ALL
    SELECT (day + time '11:30') AT TIME ZONE 'Asia/Bangkok', (day + time '12:30') AT TIME ZONE 'Asia/Bangkok', 5 FROM next_monday
)
INSERT INTO classes (trainer_id, starts_at, ends_at, capacity)
SELECT 1001, s.start_at, s.end_at, s.capacity
FROM sample s
WHERE NOT EXISTS (
    SELECT 1 FROM classes c WHERE c.trainer_id = 1001 AND c.starts_at = s.start_at AND c.ends_at = s.end_at
);
