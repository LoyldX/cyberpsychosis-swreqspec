-- ข้อมูลเทรนเนอร์ตัวอย่างสำหรับ dev เท่านั้น (004 ASM-09: เตรียมไว้ล่วงหน้า, plan D9) — ไม่ใช่ migration
-- รันด้วยมือ: psql "$DATABASE_URL" -f backend/seeds/004_trainers.sql  (รันซ้ำได้)
-- กำหนด id คงที่ 1001, 1002 เพื่อให้ข้อมูลคลาสตัวอย่างของ 005 (backend/seeds/005_classes.sql) อ้างถึงได้
-- หมายเหตุ: test ของ backend ล้างตาราง trainers ในฐานข้อมูล dev (plan D8) ต้องรันไฟล์นี้ใหม่หลังรัน test
INSERT INTO trainers (id, name) OVERRIDING SYSTEM VALUE
VALUES (1001, 'ครูเอ (ตัวอย่าง)'),
       (1002, 'ครูบี (ตัวอย่าง)')
ON CONFLICT (id) DO NOTHING;
