# Tasks: จองคลาสออกกำลังกาย

- Feature: จองคลาสออกกำลังกาย (005-class-booking)
- Spec ID: FR-BKG-01, FR-BKG-02, NFR-PERF-01, DOM-BKG-01, CON-DB-01, CON-TECH-01, CON-TECH-02
- อ้างอิง: [spec.md](./spec.md) (Draft v2 ผ่าน /clarify รอบ 2), [plan.md](./plan.md) (ร่าง 12 พฤศจิกายน 2569)
- วันที่: 12 พฤศจิกายน 2569 (แทนที่ tasks ฉบับ 11 พฤศจิกายน 2569 ที่ทุก task รอคำถามซึ่งปิดแล้ว)

มีทั้งหมด 13 task (อัปเดต 12 พ.ย. 2569: Q-01, Q-02 ของ 004 ปิดแล้ว T-10..T-12 จึงเปลี่ยนเป็นพร้อมทำ รอ task ของ 004 ตามช่อง "ต้องทำหลัง"): ร่างแรกพร้อมทำ 10 task (T-01 ถึง T-09, T-13) ครอบคลุม schema, กฎ, store, API, test ทุก AC และหน้าเว็บด้วย API จำลอง
รอ Open Questions 3 task (T-10 ถึง T-12) ทั้งหมดรอ Q-01, Q-02 ของ spec 004 (รูปแบบเวลาเข้างาน) เพราะ 005 อ่านกะจาก 004 (ASM-11) ส่วน Q-04 ของ 005 ไม่มี task เพราะอยู่ใน Out of scope

## รายการ task

### T-01 สร้าง migration ตาราง classes และ bookings
- รองรับ: CON-DB-01, FR-BKG-01, FR-BKG-02 (ASM-03, ASM-04, ASM-05, ASM-06)
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-03
- ไฟล์ที่แตะ: `backend/migrations/0004_create_class_booking.up.sql`, `backend/migrations/0004_create_class_booking.down.sql`, `backend/internal/booking/migration_test.go`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: `TestMigration0004CreatesClassBooking` ผ่านกับ PostgreSQL จริง (CHECK `ends_at > starts_at` และ `capacity > 0` ทำงาน, UNIQUE `(class_id, member_id)` ทำงาน, `member_id` ต้องมีอยู่ใน `members`, migrate down กลับ version 3 แล้วตารางหายไป)
- สถานะ: เสร็จ รอทีมตรวจ

### T-02 เขียนกฎช่วงเวลาทับซ้อนและอยู่ในกะ
- รองรับ: FR-BKG-02, DOM-BKG-01 (ASM-12, ASM-13)
- ตรวจด้วย: AC-BKG-02, AC-BKG-03, AC-BKG-04 (ระดับกฎ: `TestAC_BKG_02_Rules`, `TestAC_BKG_03_Rules`, `TestAC_BKG_04_Rules`)
- ไฟล์ที่แตะ: `backend/internal/booking/model.go`, `backend/internal/booking/model_test.go`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: test กฎทั้ง 3 ผ่าน — ทับซ้อนแบบช่วงปิด (ต่อกันพอดีนับว่าทับ), อยู่ในกะเมื่อเริ่มไม่ก่อนต้นกะและจบไม่เกินสิ้นสุดกะ (ขอบพอดีนับว่าอยู่ในกะ) และมีค่าคงที่รหัสเหตุผลครบตาม plan ข้อ 4
- สถานะ: เสร็จ รอทีมตรวจ

### T-03 สร้าง store ของคลาสและการจอง
- รองรับ: FR-BKG-01, FR-BKG-02, CON-DB-01, NFR-PERF-01 (ASM-05, ASM-09, ASM-10, plan D2, D3)
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-04 ถึง T-07 (test ระดับ store: คลาสเต็ม, จองซ้ำ, คลาสเริ่มแล้ว, ไม่พบคลาส, ทับซ้อนนับเฉพาะคลาสที่มีผู้จอง)
- ไฟล์ที่แตะ: `backend/internal/booking/store.go`, `backend/internal/booking/store_test.go`
- ต้องทำหลัง: T-01, T-02
- เสร็จเมื่อ: test store กับ PostgreSQL จริงผ่าน — `ListUpcoming` คืนเฉพาะคลาสที่ยังไม่เริ่ม, `GetForMember` คืนเฉพาะการจองของสมาชิกนั้น, `Book` ทำใน transaction เดียวพร้อม advisory lock ต่อ `trainer_id` ตรวจตามลำดับ D3 (รับตัวตรวจกะเป็นพารามิเตอร์) และเมื่อไม่ผ่านไม่มีแถวใหม่ใน `bookings`
- สถานะ: เสร็จ รอทีมตรวจ

### T-04 สร้าง service พร้อม Clock และ ShiftChecker
- รองรับ: FR-BKG-02, DOM-BKG-01 (ASM-08, ASM-11, plan D6)
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-05, T-06 (test ระดับ service: นอกกะถูกปฏิเสธ, ไม่มีข้อมูลกะถูกปฏิเสธ, เวลาปัจจุบันมาจาก Clock)
- ไฟล์ที่แตะ: `backend/internal/booking/service.go`, `backend/internal/booking/service_test.go`
- ต้องทำหลัง: T-03
- เสร็จเมื่อ: test service ผ่าน — interface `ShiftChecker` และ `Clock` ใส่ตัวจำลองได้, `ShiftChecker` ตอบ "ไม่อยู่ในกะ" หรือไม่มีข้อมูลกะได้ `ErrTrainerOffShift` และไม่มีแถวใหม่
- สถานะ: เสร็จ รอทีมตรวจ

### T-05 สร้าง Gin handler และ route ของการจอง
- รองรับ: FR-BKG-01, CON-TECH-01 (ASM-05, ASM-06, ASM-07, plan D4, D5)
- ตรวจด้วย: AC-BKG-01 (`TestAC_BKG_01_HTTP`)
- ไฟล์ที่แตะ: `backend/internal/booking/handler.go`, `backend/internal/booking/routes.go`, `backend/internal/booking/handler_test.go`
- ต้องทำหลัง: T-04
- เสร็จเมื่อ: `TestAC_BKG_01_HTTP` ผ่าน (`POST /api/bookings` → `201` แล้ว `GET /api/bookings/:id` → `200` ตรงกัน) และ test เพิ่มผ่าน: `GET /api/classes` คืนเฉพาะคลาสที่ยังไม่เริ่ม, `400` / `404` / `409` ตามรหัสใน plan ข้อ 4, `GET /api/bookings/:id` ของคนอื่นได้ `404`, ไม่มี session ได้ `401`, store ล่มได้ `503 unavailable`
- สถานะ: เสร็จ รอทีมตรวจ

### T-06 ทดสอบการบล็อกผ่าน API ตาม AC-BKG-02/03/04
- รองรับ: FR-BKG-02, DOM-BKG-01 (ASM-08, ASM-10, ASM-12, ASM-13)
- ตรวจด้วย: AC-BKG-02, AC-BKG-03, AC-BKG-04 (`TestAC_BKG_02_HTTP`, `TestAC_BKG_03_HTTP`, `TestAC_BKG_04_HTTP`)
- ไฟล์ที่แตะ: `backend/internal/booking/block_test.go`
- ต้องทำหลัง: T-05
- เสร็จเมื่อ: test ทั้ง 3 ผ่านกับ PostgreSQL จริง — ทับซ้อนกับคลาสที่มีผู้จองได้ `409 trainer_busy` (รวมต่อกันพอดี) แต่ทับกับคลาสที่ยังไม่มีผู้จองจองได้, นอกกะ (เริ่มก่อน 09:00 / จบหลัง 12:00 กับ `ShiftChecker` จำลองกะ 09:00–12:00) ได้ `409 trainer_off_shift` และทุกกรณีที่ถูกบล็อกไม่มีแถวใหม่
- สถานะ: เสร็จ รอทีมตรวจ

### T-07 ทดสอบการจองพร้อมกันตาม AC-BKG-05
- รองรับ: NFR-PERF-01, CON-DB-01 (ASM-02, plan D2)
- ตรวจด้วย: AC-BKG-05 (`TestAC_BKG_05`)
- ไฟล์ที่แตะ: `backend/internal/booking/concurrency_test.go`
- ต้องทำหลัง: T-05
- เสร็จเมื่อ: `go test -race -run TestAC_BKG_05 ./internal/booking/` ผ่านกับ PostgreSQL จริง 20 รอบติด — ทุกรอบได้ `201` หนึ่งรายการ `409 trainer_busy` หนึ่งรายการ มี 1 แถวใน `bookings` และทั้งสองคำขอได้ผลภายใน 1 วินาที
- สถานะ: เสร็จ รอทีมตรวจ

### T-08 สร้างตัวเรียก API การจองฝั่งเว็บและ API จำลอง
- รองรับ: FR-BKG-01, CON-TECH-02
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-09
- ไฟล์ที่แตะ: `frontend/src/features/booking/types.ts`, `frontend/src/features/booking/api.ts`, `frontend/src/features/booking/api.test.ts`, `frontend/src/mocks/handlers.ts`
- ต้องทำหลัง: ไม่มี (ใช้สัญญา API ใน plan ข้อ 4)
- เสร็จเมื่อ: vitest ของ `api.test.ts` ผ่าน — `fetchClasses`, `createBooking`, `fetchBooking` ส่ง `Authorization: Bearer <token>` จาก `readToken()` และแปลงผล `201` / `409` (รหัสเหตุผล) / `503` ได้ถูกต้อง และ msw มี handler จำลองทั้ง 3 endpoint
- สถานะ: เสร็จ รอทีมตรวจ

### T-09 สร้างหน้าเว็บจองคลาส /classes
- รองรับ: FR-BKG-01, CON-TECH-02 (ASM-05, plan D7)
- ตรวจด้วย: AC-BKG-01 (vitest `AC_BKG_01`)
- ไฟล์ที่แตะ: `frontend/src/features/booking/ClassBookingPage.tsx`, `frontend/src/features/booking/ClassBookingPage.test.tsx`, `frontend/src/App.tsx`
- ต้องทำหลัง: T-08
- เสร็จเมื่อ: vitest ผ่าน — `AC_BKG_01` (เห็นคลาส กดจอง แสดงการจองที่ได้), กรณี `409` แสดงข้อความตามรหัสเหตุผล, กรณี `503` แสดงข้อความให้ลองใหม่, route `/classes` อยู่ใน `RequireAuth` และ `npx tsc --noEmit` ผ่าน
- สถานะ: เสร็จ รอทีมตรวจ

### T-10 สร้างตัวอ่านเวลาเข้างานจากตารางของ 004
- รองรับ: FR-BKG-02, DOM-BKG-01 (ASM-11, ASM-13, plan D6)
- ตรวจด้วย: AC-BKG-03, AC-BKG-04 (`TestAC_BKG_03_Roster`, `TestAC_BKG_04_Roster`)
- ไฟล์ที่แตะ: `backend/internal/booking/roster_shift.go`, `backend/internal/booking/roster_shift_test.go`
- ต้องทำหลัง: T-04 และ T-05 ของ spec 004 (`roster.Covers`)
- เสร็จเมื่อ: ตัวอ่านที่ทำตาม interface `ShiftChecker` (ห่อ `roster.Covers` ของ 004) อ่านเวลาเข้างานจริงจาก PostgreSQL แล้ว test ทั้ง 2 ผ่าน (กะ 09:00–12:00: คลาสเริ่มก่อน 09:00 และคลาส 11:30–12:30 ไม่อยู่ในกะ, คลาส 09:00–10:00 และ 11:00–12:00 อยู่ในกะ)
- สถานะ: เสร็จ รอทีมตรวจ

### T-11 ประกอบการจองเข้ากับ API server
- รองรับ: CON-TECH-01, FR-BKG-01, FR-BKG-02
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-12
- ไฟล์ที่แตะ: `backend/cmd/api/main.go`
- ต้องทำหลัง: T-05, T-10
- เสร็จเมื่อ: `go build ./...` และ `go vet ./...` ผ่าน และเมื่อเปิดเซิร์ฟเวอร์กับ PostgreSQL/Redis จริง `GET /api/classes` ที่มี token ถูกต้องได้ `200` และไม่มี token ได้ `401`
- สถานะ: เสร็จ รอทีมตรวจ

### T-12 ต่อหน้าเว็บกับ API จริงและตรวจด้วยมือ
- รองรับ: FR-BKG-01, FR-BKG-02, CON-TECH-02 (ASM-04, plan D9)
- ตรวจด้วย: AC-BKG-01, AC-BKG-02 (ตรวจด้วยมือบนเบราว์เซอร์)
- ไฟล์ที่แตะ: `backend/seeds/005_classes.sql`, `docs/test_report/005-class-booking/manual_test.md`
- ต้องทำหลัง: T-09, T-11 และ T-11 ของ spec 004 (ข้อมูลเทรนเนอร์ตัวอย่าง)
- เสร็จเมื่อ: มีข้อมูลคลาสตัวอย่างที่ใช้ `trainer_id` ตรงกับข้อมูลของ 004 และขั้นตอนตรวจด้วยมือใน `manual_test.md` (ล็อกอิน → `/classes` → จองสำเร็จ; จองคลาสที่ทับกับคลาสที่มีผู้จองแล้วเห็นข้อความเทรนเนอร์ไม่ว่าง) ถูกตรวจและบันทึกผล
- สถานะ: พร้อมทำ (ต้องรอ T-11 และ 004 T-11 เสร็จก่อน)

### T-13 เขียนสัญญา API ของ 005 ใน docs/API.md
- รองรับ: FR-BKG-01, FR-BKG-02 (README ขั้นที่ 5)
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นเอกสารของ T-05 ถึง T-07
- ไฟล์ที่แตะ: `docs/API.md`
- ต้องทำหลัง: T-05, T-06, T-07
- เสร็จเมื่อ: `docs/API.md` มีหัวข้อ 005 ครบ 3 endpoint (input, output, รหัสสถานะ, รหัสเหตุผล 409) ตรงกับโค้ด และระบุชัดว่ายังไม่ได้ประกอบเข้า `main.go` จนกว่า T-11 จะเสร็จ
- สถานะ: เสร็จ รอทีมตรวจ

## ตารางตรวจความครบ

### 1. AC → task ที่ตรวจ
| AC ID | task ที่ตรวจ AC นี้ |
|---|---|
| AC-BKG-01 | T-05 (`TestAC_BKG_01_HTTP`), T-09 (`AC_BKG_01`), T-12 (ตรวจด้วยมือ) |
| AC-BKG-02 | T-02 (`TestAC_BKG_02_Rules`), T-06 (`TestAC_BKG_02_HTTP`), T-12 (ตรวจด้วยมือ) |
| AC-BKG-03 | T-02 (`TestAC_BKG_03_Rules`), T-06 (`TestAC_BKG_03_HTTP`), T-10 (`TestAC_BKG_03_Roster`) |
| AC-BKG-04 | T-02 (`TestAC_BKG_04_Rules`), T-06 (`TestAC_BKG_04_HTTP`), T-10 (`TestAC_BKG_04_Roster`) |
| AC-BKG-05 | T-07 (`TestAC_BKG_05`) |

### 2. Constraint → task ที่ทำให้เป็นจริง
| Constraint ID | task ที่ทำให้เป็นจริง |
|---|---|
| CON-DB-01 | T-01, T-03, T-07 |
| CON-TECH-01 | T-05, T-11 |
| CON-TECH-02 | T-08, T-09, T-12 |
| DOM-BKG-01 | T-02, T-03, T-04, T-06, T-10 |
| IF | ไม่มีใน spec นี้ |

## สิ่งที่ยังไม่ทำ
| รหัส | Open Question | task ที่รอ |
|---|---|---|
| Q-04 (spec 005) | สมาชิกยกเลิกการจองและดูรายการจองของตนเองได้หรือไม่ — ยังเปิด ทีมต้องการ แต่ต้องเสนอ FR ใหม่เข้า SRS ก่อน | ไม่มี task (อยู่ใน Out of scope ห้ามสร้าง task) |
| Q-01 (spec 004) | รูปแบบตารางงาน — ปิดแล้ว (004 ASM-03 ซ้ำทุกสัปดาห์) | ไม่มี task รอแล้ว |
| Q-02 (spec 004) | เวลาเข้างาน/เวลาสอน — ปิดแล้ว (004 ASM-04 แยกกัน) | ไม่มี task รอแล้ว |
| Q-01..Q-03, Q-05..Q-14 (spec 005) | ทีมตอบแล้วเป็น ASM-02..ASM-13 ชั่วคราว รอยืนยันกับผู้เกี่ยวข้อง | ไม่มี task รอ แต่ถ้าคำตอบจริงต่างไป task ที่อ้าง ASM นั้นต้องแก้ |
