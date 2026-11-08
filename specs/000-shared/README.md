# 000-shared — การตัดสินใจร่วมของทุกฟีเจอร์ (Sprint 1)

Source: `docs/srs/srs.md` v5.1 §7 | Status: Draft v1 | Owner: _(รอมอบหมาย)_ | อัปเดต: 8 พฤศจิกายน 2569

> โฟลเดอร์นี้ **ไม่ใช่ feature spec** (ไม่มี `spec.md` จึงไม่ถูกนับเป็นฟีเจอร์ตอนรัน `/clarify`) `/plan` และ `/tasks` อ่านโฟลเดอร์นี้ทุกครั้งที่มี
> ถ้าขัดกับ `docs/srs/srs.md` ให้ยึด SRS แล้วแก้ที่นี่ และ `docs/specs.md` §2 สรุปเรื่องเดียวกัน

## 1. Constraints ร่วม (ทุก plan.md ต้องอยู่ในกรอบนี้)

| ID | ข้อกำหนด (ตาม SRS §7) | ใช้กับ | สถานะ |
|---|---|---|---|
| CON-TECH-01 | ใช้ภาษา Go และ Gin framework พัฒนา Backend API | ทุกฟีเจอร์ | เลือกแล้ว |
| CON-TECH-02 | ใช้ React (Vite) เป็นเว็บแอปพลิเคชันอย่างเดียว ครอบคลุม Staff Portal, BI Dashboard และหน้าของสมาชิก (ไม่มีแอปมือถือเนทีฟ) | ทุกฟีเจอร์ | เลือกแล้ว |
| CON-DB-01 | ใช้ PostgreSQL เก็บข้อมูลสมาชิก ตารางงานเทรนเนอร์ การจองคลาส และ Inventory | 001, 004, 005, 006 | เลือกแล้ว (ตอบ Q4) |
| CON-CACHE-01 | ใช้ Redis บริหาร Session Login | 001 | ชัดเจน |
| IF-GATE-01 | เชื่อมสัญญาณผ่านเข้า-ออกจากประตูทางเข้ายิม | 007 | ประเภทสัญญาณและโปรโตคอลยังไม่ระบุ |

## 2. เทคโนโลยีที่ยังไม่ได้เลือก

SRS ไม่ได้กำหนด ทุกแถวด้านล่าง **"ทีมเลือกเอง ไม่ได้มาจาก spec"** ทีมมอบให้ AI เลือกเมื่อ 8 พฤศจิกายน 2569 (**รอทีมตรวจ**) ถ้าทีมไม่เห็นด้วยให้แก้ที่นี่ที่เดียว แล้ว plan ทุกฟีเจอร์อ้างกลับมา

| หัวข้อ | ผลเลือก | เหตุผล |
|---|---|---|
| โครงโฟลเดอร์ repo (ระดับบนสุด) | `backend/` (Go module) และ `frontend/` (Vite) | ตรงกับ CON-TECH-01/02 และกับ `docker-compose.yml` เดิม โครงภายในให้แต่ละ plan กำหนด |
| การเข้าถึงฐานข้อมูล | `pgx/v5` (ใช้ `pgxpool`) เขียน SQL ตรง ไม่ใช้ ORM | NFR-DATA-01 และ AC-BKG-05 ต้องควบคุม transaction/การล็อกเอง ORM ซ่อนรายละเอียดนี้ |
| Migration | `golang-migrate` ไฟล์ `.sql` ใน `backend/migrations/` | schema อยู่ในไฟล์ที่ตรวจ diff ได้ และรันซ้ำในการทดสอบได้ |
| Redis client (ใช้ใน 001) | `go-redis/v9` | ไลบรารีอย่างเป็นทางการของ Redis สำหรับ Go |
| Validation/HTTP | ใช้ความสามารถของ Gin (binding) | ไม่เพิ่มไลบรารี |
| Router ฝั่ง React | `react-router` | มาตรฐานของ React SPA |
| การเรียก API และ state ฝั่ง React | `fetch` + `@tanstack/react-query` สถานะหน้าจอใช้ state ของ React | ไม่เพิ่ม state manager ที่ spec ไม่ต้องการ |
| Test ฝั่ง React | `vitest` + `@testing-library/react` + `msw` (จำลอง API ตามสัญญาใน plan §4) | `/tasks` ให้ทำหน้าจอด้วย API จำลองก่อน |
| Test ฝั่ง Go | ดูหัวข้อ 3 | |

## 3. Test ฝั่ง Backend (Go)

ทุก AC ที่เป็นพฤติกรรมของ backend ต้องมี test อัตโนมัติ ชื่อมี AC ID (ดูหัวข้อ 4)

| ประเภท | เครื่องมือ | ใช้เมื่อ |
|---|---|---|
| Unit test | `testing` ของ Go + `github.com/stretchr/testify` (assert/require) | logic ล้วน เช่น คำนวณช่วงเวลาทับซ้อน (AC-BKG-02..04), การคำนวณจำนวนคน (AC-BI-02/03) |
| HTTP handler test | `net/http/httptest` + `gin.SetMode(gin.TestMode)` | endpoint ของ Gin ทุกตัว ตรวจสถานะและผลลัพธ์ที่สังเกตได้ตาม Then ของ AC |
| Integration test | PostgreSQL 16 และ Redis 7 **จริง** จาก `docker-compose.yml` (service `postgres`, `redis`) รัน migration ก่อนแต่ละชุด test และล้างข้อมูลหลังจบ | AC ที่ต้องพิสูจน์ transaction/การเก็บข้อมูล: AC-AUTH-05 (Redis), AC-INV-14 (ยอดไม่ติดลบ), AC-BKG-05 (จองพร้อมกัน) **ห้ามใช้ mock แทนฐานข้อมูล** เพราะ mock พิสูจน์ ACID ไม่ได้ |
| Concurrency test | goroutine + `sync.WaitGroup` และรัน `go test -race` | AC-BKG-05, AC-INV-14 |
| ตัวจำลองสัญญาณประตู | โค้ดทดสอบที่ส่งเหตุการณ์ผ่านเข้า/ออกเข้าระบบ | AC-BI-02..04 จนกว่าจะรู้โปรโตคอลของ IF-GATE-01 |

- **คำสั่ง:** `go test -race ./...` (integration test ต้องเปิด `docker compose up -d postgres redis` ก่อน)
- **ที่วางไฟล์:** `*_test.go` อยู่ใน package เดียวกับโค้ดที่ทดสอบ
- **Coverage:** ไม่กำหนดตัวเลขขั้นต่ำ เพราะ spec ไม่ได้กำหนด ใช้ "ทุก AC มี test" เป็นเกณฑ์
- **ข้อควรระวังของ `docker-compose.yml`:** บรรทัด mount `./backend/db/schema.sql` ชี้ไปไฟล์ที่ไม่มีแล้ว (โฟลเดอร์ `backend/` เดิมถูกลบ) ต้องเอาออกหรือเปลี่ยนเป็นการรัน `golang-migrate` ตอนสร้างโปรเจกต์ backend ใหม่ เพราะ task ที่แตะ `docker-compose.yml` ต้องมีใน `tasks.md`

## 4. ธรรมเนียมร่วม

- **ID:** ใช้ ID เดิมจาก SRS ทุกตัว ID ใหม่ที่สร้างได้มีเพียง `AC-*` (ใน spec) และ `T-xx` (ใน tasks.md) ตามวิชา ห้ามสร้าง ID อื่น
- **ชื่อ test ตาม AC ID:** ตัวอย่างของวิชา `test_AC_BKG_01` เป็นแบบ Python แต่ Go บังคับให้ชื่อฟังก์ชัน test ขึ้นต้นด้วย `Test` จึงใช้ `TestAC_BKG_01` ฝั่ง React ให้ชื่อ test มี `AC_BKG_01` (ทีมเลือกเอง ไม่ได้มาจาก spec) เพื่อให้ค้นหาตาม AC ID ได้
- **คอมเมนต์:** ฟังก์ชันหรือคอมโพเนนต์หลักใส่ 1 บรรทัดว่ารองรับ FR/CON/IF ข้อใด (AGENTS.md ข้อ 7)
- **ขอบเขต:** ห้ามสร้างสิ่งที่อยู่ใน Out of scope ของ spec ห้ามเพิ่ม requirement ที่ spec ไม่มี

## 5. คำถามเปิดร่วม

| คำถาม | กระทบ | ผู้ตอบที่น่าจะได้ |
|---|---|---|
| SRS ระบุช่องทาง "Member Mobile App" แต่ CON-TECH-02 กำหนดเป็นเว็บอย่างเดียว ทีมยืนยันหรือไม่ว่าหน้าสมาชิกส่งมอบเป็นเว็บแบบ responsive (กระทบการตีความ UI-AUTH-01 "เปิดแอป") | 001, 005 | Product Owner (เสนอ) |
| เจ้าหน้าที่/ผู้จัดการล็อกอินอย่างไร (SRS ไม่กำหนด) | 004, 006, 007 | Product Owner (เสนอ) |
| พฤติกรรมเมื่อ Redis, PostgreSQL หรือประตูไม่ตอบสนอง | ทุกฟีเจอร์ | Tech Lead (เสนอ) |

รายละเอียดคำถามเปิดต่อฟีเจอร์อยู่ใน `docs/specs.md` §3–§4
