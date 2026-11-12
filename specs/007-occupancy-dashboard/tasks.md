# Tasks: Occupancy Dashboard

- Feature: Occupancy Dashboard
- Spec ID: FR-BI-01, NFR-PERF-02, DOM-CAP-01, IF-GATE-01, CON-TECH-01, CON-TECH-02
- อ้างอิง: [spec.md](./spec.md) (Draft v2), [plan.md](./plan.md)
- วันที่: 12 พฤศจิกายน 2569

มีทั้งหมด 12 task แบ่งเป็น backend 7 task (T-01..T-07), frontend 2 task (T-08, T-09), ต่อ API จริง 1 task (T-10), เอกสาร API 1 task (T-11) และ adapter ประตูจริง 1 task (T-12)
มี 1 task ที่รอ Open Question คือ T-12 (รอ Q-01); task อื่นใช้ ASM-01..07 ของทีมซึ่งยังต้องยืนยันกับผู้เกี่ยวข้องจริง (Q3, Q-02..Q-07)

## รายการ task

### T-01 สร้าง migration ตาราง gate_signals
- รองรับ: DOM-CAP-01, FR-BI-01, IF-GATE-01 (ASM-02, ASM-04)
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-03
- ไฟล์ที่แตะ: `backend/migrations/0003_create_gate_signals.up.sql`, `backend/migrations/0003_create_gate_signals.down.sql`, `backend/internal/occupancy/migration_test.go`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: test ใน `migration_test.go` รัน migrate up ได้ตาราง `gate_signals` ตาม plan §3 ฐานข้อมูลปฏิเสธ `direction` อื่นนอกจาก `in`/`out` และ `occupancy_after < 0` ส่วน migrate down ลบตารางได้
- สถานะ: โค้ดเสร็จ รอรัน test กับ PostgreSQL (เครื่องที่ implement ไม่มี PostgreSQL/Docker; ทีมสั่งให้เขียนโค้ดไปก่อน)

### T-02 เขียนกฎคำนวณจำนวนและจุดเริ่มวัน
- รองรับ: DOM-CAP-01 (ASM-01, ASM-03, ASM-04)
- ตรวจด้วย: AC-BI-02, AC-BI-03, AC-BI-05 (ส่วน unit ของกฎคำนวณ)
- ไฟล์ที่แตะ: `backend/internal/occupancy/model.go`, `backend/internal/occupancy/model_test.go`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: `TestAC_BI_02`, `TestAC_BI_03`, `TestAC_BI_05` ระดับ unit ผ่าน (`ApplySignal(10,in)=11`, `ApplySignal(10,out)=9`, `ApplySignal(0,out)=0`) และ test ของ `StartOfDay` ยืนยันว่าเวลา 23:59 และ 00:00 Asia/Bangkok อยู่คนละวัน
- สถานะ: เสร็จ รอทีมตรวจ

### T-03 สร้าง store บันทึกสัญญาณและอ่านจำนวนปัจจุบัน
- รองรับ: FR-BI-01, DOM-CAP-01, IF-GATE-01 (ASM-02, ASM-04, ASM-05)
- ตรวจด้วย: AC-BI-02 (กรณีสัญญาณพร้อมกัน), AC-BI-05, AC-BI-07 (ระดับ store)
- ไฟล์ที่แตะ: `backend/internal/occupancy/store.go`, `backend/internal/occupancy/store_test.go`
- ต้องทำหลัง: T-01, T-02
- เสร็จเมื่อ: integration test บน PostgreSQL จริงผ่านด้วย `go test -race` ครบ 3 ข้อ
  - `RecordSignal` ใช้ transaction + `pg_advisory_xact_lock` บันทึก `occupancy_after` ถูกต้อง
  - สัญญาณ `in` พร้อมกันหลาย goroutine ได้จำนวนเท่ากับจำนวนสัญญาณ
  - store ตัวใหม่บนฐานข้อมูลเดิมอ่าน `CurrentCount` ได้ค่าเดิม
- สถานะ: โค้ดเสร็จ รอรัน test กับ PostgreSQL

### T-04 สร้าง broadcaster และ service พร้อมนาฬิกาจำลอง
- รองรับ: NFR-PERF-02, DOM-CAP-01 (ASM-03, ASM-06)
- ตรวจด้วย: AC-BI-06
- ไฟล์ที่แตะ: `backend/internal/occupancy/broadcaster.go`, `backend/internal/occupancy/service.go`, `backend/internal/occupancy/service_test.go`
- ต้องทำหลัง: T-03
- เสร็จเมื่อ: `TestAC_BI_06` ผ่าน: ใช้นาฬิกาจำลอง จำนวน 10 ตอน 23:59 แล้วเลื่อนไป 00:00 Asia/Bangkok ได้ `CurrentCount` = 0 และผู้ฟังได้รับค่า `count: 0`; บันทึกสัญญาณผ่าน service แล้วผู้ฟังทุกรายได้รับจำนวนใหม่
- สถานะ: โค้ดเสร็จ รอรัน test กับ PostgreSQL (test ส่วนที่ไม่ใช้ DB ผ่านแล้ว)

### T-05 สร้าง API อ่านจำนวนและ endpoint ตัวจำลองสัญญาณ
- รองรับ: FR-BI-01, IF-GATE-01 (ตัวจำลองเท่านั้น), DOM-CAP-01, CON-TECH-01 (ASM-07)
- ตรวจด้วย: AC-BI-01, AC-BI-02, AC-BI-03, AC-BI-05, AC-BI-07 (ระดับ HTTP)
- ไฟล์ที่แตะ: `backend/internal/occupancy/handler.go`, `backend/internal/occupancy/routes.go`, `backend/internal/occupancy/handler_test.go`
- ต้องทำหลัง: T-04
- เสร็จเมื่อ: `TestAC_BI_01`, `TestAC_BI_02`, `TestAC_BI_03`, `TestAC_BI_05`, `TestAC_BI_07` ระดับ HTTP ผ่านด้วย `httptest` + PostgreSQL จริง
  - `POST /api/gate/signals` ด้วย `direction` ผิดได้ `400`
  - เมื่อ DB ไม่ตอบได้ `503` (D6)
- สถานะ: โค้ดเสร็จ รอรัน test กับ PostgreSQL (test ส่วนที่ไม่ใช้ DB ผ่านแล้ว)

### T-06 สร้าง SSE endpoint สำหรับอัปเดตจำนวน
- รองรับ: NFR-PERF-02, CON-TECH-01 (ASM-06)
- ตรวจด้วย: AC-BI-04 (ฝั่ง backend)
- ไฟล์ที่แตะ: `backend/internal/occupancy/handler.go`, `backend/internal/occupancy/routes.go`, `backend/internal/occupancy/stream_test.go`
- ต้องทำหลัง: T-05
- เสร็จเมื่อ: `TestAC_BI_04` ผ่าน: เปิด `GET /api/occupancy/stream` ผ่าน `httptest.Server` ได้ event แรกเป็นจำนวนปัจจุบัน และหลัง `POST /api/gate/signals` ตอบกลับ ได้ event ที่มีจำนวนใหม่ภายใน 1 วินาที; ปิดการเชื่อมต่อแล้วผู้ฟังถูกถอดออกจาก broadcaster
- สถานะ: พร้อมทำ

### T-07 ประกอบ occupancy เข้ากับเซิร์ฟเวอร์
- รองรับ: CON-TECH-01, FR-BI-01, NFR-PERF-02
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-10
- ไฟล์ที่แตะ: `backend/cmd/api/main.go`
- ต้องทำหลัง: T-06
- เสร็จเมื่อ: `go build ./...` และ `go vet ./...` ผ่าน เปิดเซิร์ฟเวอร์แล้ว `GET /api/occupancy` ตอบ `200` และตัวตั้งเวลา 00:00 หยุดเมื่อเซิร์ฟเวอร์ปิด
- สถานะ: พร้อมทำ

### T-08 สร้าง type, ตัวเรียก API และ handler จำลองฝั่ง frontend
- รองรับ: FR-BI-01, NFR-PERF-02, CON-TECH-02
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-09
- ไฟล์ที่แตะ: `frontend/src/features/occupancy/types.ts`, `frontend/src/features/occupancy/api.ts`, `frontend/src/features/occupancy/api.test.ts`, `frontend/src/mocks/handlers.ts`
- ต้องทำหลัง: ไม่มี (ใช้ API จำลองตามสัญญาใน plan §4.1)
- เสร็จเมื่อ: test ใน `api.test.ts` ผ่าน: `fetchOccupancy()` อ่าน `{count, as_of}` จาก msw ได้ และ `subscribeOccupancy()` เรียก callback เมื่อ `EventSource` จำลองส่ง event `occupancy` และคืนฟังก์ชันที่ปิดการเชื่อมต่อ
- สถานะ: พร้อมทำ

### T-09 สร้างหน้า Dashboard และ route /dashboard
- รองรับ: FR-BI-01, NFR-PERF-02, CON-TECH-02 (ASM-07)
- ตรวจด้วย: AC-BI-01, AC-BI-04 (ฝั่ง frontend)
- ไฟล์ที่แตะ: `frontend/src/features/occupancy/OccupancyDashboardPage.tsx`, `frontend/src/features/occupancy/OccupancyDashboardPage.test.tsx`, `frontend/src/App.tsx`
- ต้องทำหลัง: T-08
- เสร็จเมื่อ: test `AC_BI_01` และ `AC_BI_04` ผ่าน: หน้า `/dashboard` แสดง 3 เมื่อ msw ตอบ `count: 3` และเปลี่ยนเป็น 4 เมื่อ `EventSource` จำลองส่ง `count: 4` โดยไม่ remount; route `/dashboard` เปิดได้โดยไม่ต้องล็อกอิน; test เดิมของ `App` ยังผ่าน
- สถานะ: พร้อมทำ

### T-10 ต่อหน้า Dashboard กับ API จริงและตรวจด้วยมือ
- รองรับ: FR-BI-01, NFR-PERF-02, IF-GATE-01 (ตัวจำลอง), CON-TECH-01, CON-TECH-02
- ตรวจด้วย: AC-BI-01, AC-BI-04 (ตรวจด้วยมือบนเบราว์เซอร์จริง)
- ไฟล์ที่แตะ: `frontend/vite.config.ts` (เพิ่ม dev proxy `/api` → `http://localhost:8080` ตาม `defaultHTTPAddr` ใน `backend/internal/config/config.go`), `docs/test_report/007-occupancy-dashboard/manual_test.md`
- ต้องทำหลัง: T-07, T-09
- เสร็จเมื่อ: เปิด backend + `npm run dev` แล้วตรวจตามขั้นตอนใน `manual_test.md` ผ่าน
  - ยิง `POST /api/gate/signals` เข้า 5 ออก 2 แล้วหน้า `/dashboard` แสดง 3
  - เปิดหน้าค้างไว้ ยิงสัญญาณเพิ่ม 1 ครั้ง ตัวเลขเปลี่ยนเองโดยไม่รีเฟรชหน้า
- สถานะ: พร้อมทำ

### T-11 เขียนสัญญา API ของ 007 ใน docs/API.md
- รองรับ: FR-BI-01, NFR-PERF-02, IF-GATE-01 (ตัวจำลอง)
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นขั้นที่ 5 ของ `specs/README.md`
- ไฟล์ที่แตะ: `docs/API.md`
- ต้องทำหลัง: T-07
- เสร็จเมื่อ: `docs/API.md` มีหัวข้อ 007 ครบ 3 endpoint ตาม plan §4.1 (input, output, รหัสสถานะ, รูปแบบ event SSE) ระบุว่า `POST /api/gate/signals` เป็นตัวจำลองเท่านั้น (Q-01) และไม่ต้องยืนยันตัวตน (ASM-07) ตรงกับโค้ดจริง
- สถานะ: พร้อมทำ

### T-12 สร้าง adapter รับสัญญาณจากประตูจริง
- รองรับ: IF-GATE-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ (spec ยังไม่มี AC ของ IF-GATE-01 จนกว่าจะรู้โปรโตคอล)
- ไฟล์ที่แตะ: ยังกำหนดไม่ได้จนกว่าจะได้คำตอบ Q-01 (คาดว่าอยู่ใน `backend/internal/occupancy/`)
- ต้องทำหลัง: T-04
- เสร็จเมื่อ: ยังกำหนดไม่ได้จนกว่าจะได้คำตอบ Q-01 และทีมเพิ่ม AC ของ IF-GATE-01 ใน spec
- สถานะ: รอ Q-01

## ตารางตรวจความครบ

### Acceptance Criteria
| AC ID | task ที่ตรวจ AC นี้ |
|---|---|
| AC-BI-01 | T-05, T-09, T-10 |
| AC-BI-02 | T-02, T-03, T-05 |
| AC-BI-03 | T-02, T-05 |
| AC-BI-04 | T-06, T-09, T-10 |
| AC-BI-05 | T-02, T-03, T-05 |
| AC-BI-06 | T-04 |
| AC-BI-07 | T-03, T-05 |

### Constraints
| Constraint ID | task ที่ทำให้เป็นจริง |
|---|---|
| IF-GATE-01 | T-05 (ตัวจำลอง), T-10, T-12 (adapter จริง รอ Q-01) |
| DOM-CAP-01 | T-01, T-02, T-03, T-04 |
| CON-TECH-01 | T-05, T-06, T-07 |
| CON-TECH-02 | T-08, T-09, T-10 |

## สิ่งที่ยังไม่ทำ
| รหัส | Open Question จาก spec | task ที่รอ |
|---|---|---|
| Q3 (SRS) | ประตูทางเข้าบันทึกทั้งขาเข้าและขาออกหรือไม่ — ทีมใช้ ASM-01 ชั่วคราว | ไม่มี task รอ; ถ้าตอบว่าบันทึกแค่ขาเข้า ต้องแก้ T-02, T-03, T-05 |
| Q8 (SRS) | ข้อมูลลงทะเบียนเกี่ยวข้องกับการนับคนเข้ายิมและ Excel เดิมหรือไม่ — Out of scope | ไม่มี task (Out of scope) |
| Q10 (SRS) | ปัญหาเครื่องเล่นมีคนใช้เป็นประจำ ทีมต้องการให้ระบบช่วยแก้หรือไม่ — Out of scope | ไม่มี task (Out of scope) |
| Q-01 | ประตูส่งสัญญาณมาในรูปแบบใด และมีประตูกี่จุด (IF-GATE-01) — ไม่รู้ ต้องถามฝ่ายอาคารสถานที่ | **T-12 รอ Q-01** |
| Q-02 | รอบเวลาอัปเดตของ NFR-PERF-02 — ทีมใช้ภายใน 1 วินาที (ASM-06) รอยืนยัน | ไม่มี task รอ; ถ้าเปลี่ยนต้องแก้ T-06, T-09 |
| Q-03 | สัญญาณซ้ำ ผิดลำดับ หรือขาดการเชื่อมต่อ — ทีมใช้รับทุกสัญญาณตามจริง (ASM-05) รอยืนยัน | ไม่มี task รอ |
| Q-04 | ผ่านออกเมื่อจำนวนเป็น 0 — ทีมใช้คงไว้ที่ 0 (ASM-04) รอยืนยัน | ไม่มี task รอ; ถ้าเปลี่ยนต้องแก้ T-02 |
| Q-05 | เก็บเป็นประวัติหรือเฉพาะจำนวน — ทีมใช้เก็บทุกเหตุการณ์ใน PostgreSQL (ASM-02) รอยืนยัน | ไม่มี task รอ |
| Q-06 | ผู้จัดการล็อกอินอย่างไร และสิทธิ์ดู Dashboard — ทีมใช้ไม่ต้องล็อกอิน (ASM-07) | ไม่มี task (ไม่สร้างระบบล็อกอินผู้จัดการ) |
| Q-07 | จำนวนปัจจุบันเริ่มนับใหม่เมื่อใด — ทีมใช้ทุกวัน 00:00 (ASM-03) รอยืนยัน | ไม่มี task รอ; ถ้าเปลี่ยนต้องแก้ T-02, T-04 |
