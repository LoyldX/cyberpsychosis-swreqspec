# Tasks: Trainer Roster

- Feature: Trainer Roster (004-trainer-roster)
- Spec ID: FR-ROST-01, CON-DB-01, CON-TECH-01, CON-TECH-02
- อ้างอิง: [spec.md](./spec.md) (Draft v2 ผ่าน /clarify รอบ 2), [plan.md](./plan.md) (ร่าง 12 พฤศจิกายน 2569)
- วันที่: 12 พฤศจิกายน 2569 (แทนที่ tasks ฉบับ 10 พฤศจิกายน 2569 ที่ทุก task รอคำถามซึ่งปิดแล้ว)

มีทั้งหมด 12 task ทุก task พร้อมทำ (spec 004 ไม่มี Open Question ที่ยังเปิด)
T-05 (`roster.Covers`) เป็นสิ่งที่ปลดล็อก 005 T-10

## รายการ task

### T-01 สร้าง migration ตาราง trainers และ roster_entries
- รองรับ: CON-DB-01, FR-ROST-01 (ASM-01, ASM-03, ASM-04, ASM-09)
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-03
- ไฟล์ที่แตะ: `backend/migrations/0005_create_trainer_roster.up.sql`, `backend/migrations/0005_create_trainer_roster.down.sql`, `backend/internal/roster/migration_test.go`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: `TestMigration0005CreatesTrainerRoster` ผ่านกับ PostgreSQL จริง (CHECK `kind`, `weekday` 1..7, `end_time > start_time` ทำงาน, `trainer_id` ต้องมีใน `trainers`, migrate down กลับ version 4 แล้วตารางหายไป)
- สถานะ: เสร็จ รอทีมตรวจ

### T-02 เขียนกฎเวลาของตารางงาน
- รองรับ: FR-ROST-01 (ASM-02, ASM-03, ASM-04, ASM-10)
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-03 ถึง T-05
- ไฟล์ที่แตะ: `backend/internal/roster/model.go`, `backend/internal/roster/model_test.go`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: unit test ผ่าน — แปลง/ตรวจ `"HH:MM"` และวัน 1..7, เริ่มต้องก่อนจบ, ทับซ้อนแบบไม่นับต่อกันพอดี, ช่วงเวลาอยู่ในชุดเวลาเข้างาน (รวมช่วงที่ต่อกัน), แปลงช่วงเวลาจริงเป็นวัน+เวลา Asia/Bangkok (ข้ามวัน = ไม่อยู่ในกะ) และมีรหัสเหตุผลครบตาม plan ข้อ 4
- สถานะ: เสร็จ รอทีมตรวจ

### T-03 สร้าง store อ่านและเพิ่มรายการตารางงาน
- รองรับ: FR-ROST-01, CON-DB-01 (ASM-02, ASM-04, ASM-10, plan D2)
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-06 (test ระดับ store: เพิ่ม/อ่าน, ทับซ้อน, ต่อกันพอดีได้, เวลาสอนนอกเวลาเข้างาน, ไม่พบเทรนเนอร์)
- ไฟล์ที่แตะ: `backend/internal/roster/store.go`, `backend/internal/roster/store_test.go`
- ต้องทำหลัง: T-01, T-02
- เสร็จเมื่อ: test store กับ PostgreSQL จริงผ่าน — `ListTrainers`, `ListEntries` (รวมที่ยกเลิก), `Create` ใน transaction พร้อม advisory lock `(5005, trainer_id)` และไม่ผ่านกฎแล้วไม่บันทึก
- สถานะ: เสร็จ รอทีมตรวจ

### T-04 สร้าง store แก้ ยกเลิก และลบรายการตารางงาน
- รองรับ: FR-ROST-01, CON-DB-01 (ASM-01, ASM-04, ASM-05, ASM-07, Q-11)
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-06, T-07 (test ระดับ store: แก้ทีละรายการ, ยกเลิกแล้วไม่นับเป็นเวลาเข้างาน, ลบ, เวลาสอนหลุดเวลาเข้างานถูกปฏิเสธ, คลาสที่มีผู้จองและยังไม่เริ่มหลุดกะถูกปฏิเสธ, คลาสที่เริ่มแล้วไม่ขวาง)
- ไฟล์ที่แตะ: `backend/internal/roster/store_change.go`, `backend/internal/roster/store_change_test.go`
- ต้องทำหลัง: T-03
- เสร็จเมื่อ: test กับ PostgreSQL จริงผ่าน — `Update`, `Cancel`, `Delete` ทำใน transaction พร้อม advisory lock ตรวจตามลำดับ plan ข้อ 4 และไม่ผ่านแล้วไม่บันทึก
- สถานะ: เสร็จ รอทีมตรวจ

### T-05 สร้าง Covers ให้ 005 ตรวจว่าคลาสอยู่ในเวลาเข้างาน
- รองรับ: FR-ROST-01, CON-DB-01 (ASM-01, ASM-03, ASM-10; 005 ASM-11, ASM-13)
- ตรวจด้วย: ไม่มี AC ตรง ๆ ของ 004 เป็นงานพื้นฐานของ 005 T-10
- ไฟล์ที่แตะ: `backend/internal/roster/covers.go`, `backend/internal/roster/covers_test.go`
- ต้องทำหลัง: T-03
- เสร็จเมื่อ: test กับ PostgreSQL จริงผ่าน — ในกะ/นอกกะ, ขอบพอดีนับว่าอยู่ในกะ, คร่อมสองกะที่ต่อกันนับว่าอยู่ในกะ, กะที่ยกเลิกไม่นับ, ข้ามวันไม่อยู่ในกะ, เทรนเนอร์ไม่มีกะ = ไม่อยู่ในกะ
- สถานะ: เสร็จ รอทีมตรวจ

### T-06 สร้าง Gin handler และ route ของตารางงาน
- รองรับ: FR-ROST-01, CON-TECH-01 (ASM-06, ASM-08, plan D3, D5)
- ตรวจด้วย: AC-ROST-01, AC-ROST-02 (`TestAC_ROST_01_HTTP`, `TestAC_ROST_02_HTTP`)
- ไฟล์ที่แตะ: `backend/internal/roster/handler.go`, `backend/internal/roster/routes.go`, `backend/internal/roster/handler_test.go`
- ต้องทำหลัง: T-04, T-05
- เสร็จเมื่อ: `TestAC_ROST_01_HTTP` และ `TestAC_ROST_02_HTTP` ผ่าน และ test เพิ่มผ่าน: `GET /api/trainers`, `400` / `404` / `409` ตามรหัสใน plan ข้อ 4, `503 unavailable` เมื่อ store ล่ม, ไม่ต้องมี token
- สถานะ: เสร็จ รอทีมตรวจ

### T-07 ทดสอบการปรับปรุงตารางตาม AC-ROST-03/04
- รองรับ: FR-ROST-01 (ASM-01, ASM-05)
- ตรวจด้วย: AC-ROST-03, AC-ROST-04 (`TestAC_ROST_03_HTTP`, `TestAC_ROST_04_HTTP`)
- ไฟล์ที่แตะ: `backend/internal/roster/update_test.go`
- ต้องทำหลัง: T-06
- เสร็จเมื่อ: test ทั้ง 2 ผ่านกับ PostgreSQL จริง — `PATCH` แล้ว `GET` ซ้ำเห็นค่าใหม่, แก้/ยกเลิก/ลบของเทรนเนอร์ A แล้วตารางของ B เหมือนเดิมทุกฟิลด์
- สถานะ: พร้อมทำ

### T-08 สร้างตัวเรียก API ตารางงานฝั่งเว็บและ API จำลอง
- รองรับ: FR-ROST-01, CON-TECH-02
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-09
- ไฟล์ที่แตะ: `frontend/src/features/roster/types.ts`, `frontend/src/features/roster/api.ts`, `frontend/src/features/roster/api.test.ts`, `frontend/src/mocks/handlers.ts`
- ต้องทำหลัง: ไม่มี (ใช้สัญญา API ใน plan ข้อ 4)
- เสร็จเมื่อ: vitest ของ `api.test.ts` ผ่าน — เรียกทั้ง 6 endpoint และแปลงผล `2xx` / `409` (รหัสเหตุผล) / `503` ได้ถูกต้อง และ msw มี handler จำลองครบ
- สถานะ: พร้อมทำ

### T-09 สร้างหน้าเว็บจัดตารางงาน /roster
- รองรับ: FR-ROST-01, CON-TECH-02 (ASM-06, ASM-08, plan D7)
- ตรวจด้วย: AC-ROST-01 (vitest `AC_ROST_01`), AC-ROST-03 (vitest `AC_ROST_03`)
- ไฟล์ที่แตะ: `frontend/src/features/roster/RosterPage.tsx`, `frontend/src/features/roster/RosterPage.test.tsx`, `frontend/src/App.tsx`
- ต้องทำหลัง: T-08
- เสร็จเมื่อ: vitest ผ่าน — `AC_ROST_01` (เลือกเทรนเนอร์ เพิ่มเวลาเข้างาน เห็นในตาราง), `AC_ROST_03` (แก้แล้วเห็นค่าใหม่), `409` แสดงเหตุผล, `503` แจ้งลองใหม่และคงค่าในฟอร์ม, route `/roster` ไม่ครอบ `RequireAuth`, `npx tsc --noEmit` ผ่าน
- สถานะ: พร้อมทำ

### T-10 ประกอบตารางงานเข้ากับ API server
- รองรับ: CON-TECH-01, FR-ROST-01 (ASM-06)
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-11
- ไฟล์ที่แตะ: `backend/cmd/api/main.go`
- ต้องทำหลัง: T-06
- เสร็จเมื่อ: `go build ./...` และ `go vet ./...` ผ่าน และเมื่อเปิดเซิร์ฟเวอร์กับ PostgreSQL จริง `GET /api/trainers` ได้ `200` โดยไม่ต้องมี token
- สถานะ: พร้อมทำ

### T-11 เตรียมข้อมูลเทรนเนอร์ตัวอย่างและแผนตรวจด้วยมือ
- รองรับ: FR-ROST-01, CON-TECH-02 (ASM-09, plan D9)
- ตรวจด้วย: AC-ROST-01, AC-ROST-02, AC-ROST-03 (ตรวจด้วยมือบนเบราว์เซอร์)
- ไฟล์ที่แตะ: `backend/seeds/004_trainers.sql`, `docs/test_report/004-trainer-roster/manual_test.md`
- ต้องทำหลัง: T-09, T-10
- เสร็จเมื่อ: มีไฟล์ข้อมูลเทรนเนอร์ตัวอย่างที่รันได้ และขั้นตอนตรวจด้วยมือใน `manual_test.md` ถูกตรวจและบันทึกผล
- สถานะ: พร้อมทำ

### T-12 เขียนสัญญา API ของ 004 ใน docs/API.md
- รองรับ: FR-ROST-01 (README ขั้นที่ 5)
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นเอกสารของ T-06, T-07
- ไฟล์ที่แตะ: `docs/API.md`
- ต้องทำหลัง: T-06, T-07
- เสร็จเมื่อ: `docs/API.md` มีหัวข้อ 004 ครบ 6 endpoint (input, output, รหัสสถานะ, รหัสเหตุผล 409) ตรงกับโค้ด และระบุว่าไม่ต้องล็อกอิน (ASM-06)
- สถานะ: พร้อมทำ

## ตารางตรวจความครบ

### 1. AC → task ที่ตรวจ
| AC ID | task ที่ตรวจ AC นี้ |
|---|---|
| AC-ROST-01 | T-06 (`TestAC_ROST_01_HTTP`), T-09 (`AC_ROST_01`), T-11 (ตรวจด้วยมือ) |
| AC-ROST-02 | T-06 (`TestAC_ROST_02_HTTP`), T-11 (ตรวจด้วยมือ) |
| AC-ROST-03 | T-07 (`TestAC_ROST_03_HTTP`), T-09 (`AC_ROST_03`), T-11 (ตรวจด้วยมือ) |
| AC-ROST-04 | T-07 (`TestAC_ROST_04_HTTP`) |

### 2. Constraint → task ที่ทำให้เป็นจริง
| Constraint ID | task ที่ทำให้เป็นจริง |
|---|---|
| CON-DB-01 | T-01, T-03, T-04, T-05 |
| CON-TECH-01 | T-06, T-10 |
| CON-TECH-02 | T-08, T-09, T-11 |
| DOM | ไม่มีใน spec นี้ |
| IF | ไม่มีใน spec นี้ |

## สิ่งที่ยังไม่ทำ
| รหัส | Open Question | task ที่รอ |
|---|---|---|
| — | spec 004 ไม่มี Open Question ที่ยังเปิด (Q-01..Q-11 ทีมตอบแล้วเป็น ASM-01, ASM-03..ASM-10 ชั่วคราว รอยืนยันกับผู้จัดการยิม / Product Owner / Tech Lead) | ไม่มี task รอ แต่ถ้าคำตอบจริงต่างไป task ที่อ้าง ASM นั้นต้องแก้ |
