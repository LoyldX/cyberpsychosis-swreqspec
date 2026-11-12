# แผน: Occupancy Dashboard

Feature: Occupancy Dashboard | Spec: `specs/007-occupancy-dashboard/spec.md` (Draft v2, 12 พฤศจิกายน 2569) | Spec ID: FR-BI-01, NFR-PERF-02, DOM-CAP-01, IF-GATE-01 | อยู่ในกรอบ `specs/000-shared/README.md` | Status: ร่างโดย AI รอทีมตรวจ | อัปเดต: 12 พฤศจิกายน 2569

## 1. สรุปแนวทาง
ฟีเจอร์นี้ให้ผู้จัดการเห็นจำนวนผู้ใช้บริการภายในยิมขณะนี้บนหน้าเว็บ Dashboard ที่อัปเดตเอง แทนการคีย์ Excel (FR-BI-01, NFR-PERF-02)
Backend Go/Gin รับสัญญาณผ่านประตู บันทึกทุกครั้งลง PostgreSQL แล้วคำนวณจำนวนตามกฎ เข้า +1 ออก -1 ไม่ต่ำกว่า 0 และนับใหม่ทุกวัน 00:00 (Asia/Bangkok) (DOM-CAP-01, ASM-01..04)
หน้าเว็บ React (Vite) ดึงจำนวนตอนเปิดหน้า แล้วรับค่าใหม่ทาง Server-Sent Events (SSE) ทันทีที่มีสัญญาณ ภายใน 1 วินาที (NFR-PERF-02, ASM-06)
โปรโตคอลจริงของประตูยังไม่รู้ (Q-01) จึงสร้างเฉพาะ endpoint สำหรับตัวจำลองสัญญาณ ส่วน adapter ของประตูจริงยังไม่สร้าง (IF-GATE-01)
หน้า Dashboard เปิดได้โดยไม่ต้องล็อกอิน (ASM-07) ไม่สร้างระบบล็อกอินผู้จัดการ ไม่กรองสัญญาณซ้ำ และไม่แสดงสถานะประตู (Out of scope, ASM-05)

## 2. เทคโนโลยีที่ใช้
| สิ่งที่เลือก | มาจาก | หมายเหตุ |
|---|---|---|
| Go + Gin | CON-TECH-01 | Backend API และ SSE endpoint |
| React (Vite) เว็บอย่างเดียว | CON-TECH-02 | หน้า Dashboard ไม่มีแอปมือถือเนทีฟ |
| PostgreSQL 16 | ทีมเลือกเอง ไม่ได้มาจาก spec (ASM-02) | SRS §7 ไม่ผูก 007 กับ CON-DB-01 ทีมเลือกใช้ PostgreSQL ตัวเดียวกับฟีเจอร์อื่น |
| `pgx/v5` (`pgxpool`) + SQL ตรง | ทีมเลือกเอง ไม่ได้มาจาก spec | ตาม `000-shared` §2 ใช้ transaction + advisory lock ให้การคำนวณจำนวนเรียงตามลำดับที่ได้รับ (ASM-05) |
| `golang-migrate` ไฟล์ `.sql` | ทีมเลือกเอง ไม่ได้มาจาก spec | ตาม `000-shared` §2 ไฟล์ `backend/migrations/0003_*` |
| Server-Sent Events (Gin `c.SSEvent` / `c.Stream`) ฝั่ง backend และ `EventSource` ของเบราว์เซอร์ฝั่ง frontend | ทีมเลือกเอง ไม่ได้มาจาก spec | ใช้ทำ NFR-PERF-02/ASM-06 (ส่งข้อมูลทางเดียวจาก server ไม่ต้องเพิ่มไลบรารี; `gin-contrib/sse` มีใน `go.mod` อยู่แล้วแบบ indirect) ดู D1 |
| `time/tzdata` (standard library) | ทีมเลือกเอง ไม่ได้มาจาก spec | ฝังข้อมูล timezone ให้ `Asia/Bangkok` (ASM-03) ใช้ได้แม้ image ไม่มี tzdata |
| `@tanstack/react-query` + `fetch` | ทีมเลือกเอง ไม่ได้มาจาก spec | ตาม `000-shared` §2 ใช้ดึงจำนวนครั้งแรก |
| Go `testing` + `testify`, `httptest`, PostgreSQL จริงจาก `docker-compose.yml` | ทีมเลือกเอง ไม่ได้มาจาก spec | ตาม `000-shared` §3 ไม่ใช้ mock แทนฐานข้อมูลใน AC-BI-01, 05, 07 |
| `vitest` + `@testing-library/react` + `msw` + `EventSource` จำลองใน test | ทีมเลือกเอง ไม่ได้มาจาก spec | jsdom ไม่มี `EventSource` จึงใช้คลาสจำลองในไฟล์ test |

ไม่เพิ่ม dependency ใหม่ใน `go.mod` หรือ `package.json`

## 3. โมเดลข้อมูล
| Entity | ฟิลด์หลัก | รองรับ |
|---|---|---|
| `gate_signals` (ตารางใหม่ migration `0003`) | `id BIGSERIAL PK` (ลำดับที่ระบบได้รับ), `direction TEXT NOT NULL CHECK (direction IN ('in','out'))`, `received_at TIMESTAMPTZ NOT NULL DEFAULT now()`, `occupancy_after INTEGER NOT NULL CHECK (occupancy_after >= 0)`; index ที่ `received_at` | FR-BI-01, DOM-CAP-01, IF-GATE-01, ASM-02, ASM-04, ASM-05 |

กฎการคำนวณ (DOM-CAP-01):
- **จำนวนปัจจุบัน** = `occupancy_after` ของแถวล่าสุด (เรียงตาม `id`) ที่ `received_at` ≥ 00:00 Asia/Bangkok ของวันนี้ ถ้าวันนี้ยังไม่มีแถว จำนวนปัจจุบันคือ 0 (ASM-03)
- **บันทึกสัญญาณใหม่** ทำใน transaction เดียว: `pg_advisory_xact_lock` → อ่านจำนวนปัจจุบัน → คำนวณ `in` = +1, `out` = max(0, n-1) (ASM-04) → insert แถวพร้อม `occupancy_after` → commit การล็อกทำให้สัญญาณที่มาพร้อมกันถูกคำนวณทีละตัวตามลำดับที่ได้รับ (ASM-05)
- เก็บทุกสัญญาณ รวมถึงสัญญาณออกที่จำนวนเป็น 0 อยู่แล้ว (ASM-04) จำนวนคำนวณจากประวัติในฐานข้อมูล จึงยังอยู่หลังรีสตาร์ต (ASM-02, AC-BI-07)
- ไม่มีฟิลด์ระบุตัวบุคคล ไม่มี `gate_id` (จำนวนประตูยังไม่รู้ Q-01) และไม่มีนโยบายลบข้อมูลเก่า (spec ไม่กำหนด)

## 4. API / หน้าจอ

### 4.1 Backend (Gin)
| Method / path | Input → Output หลัก | รองรับ |
|---|---|---|
| `GET /api/occupancy` | ไม่มี → `200 {"count": int, "as_of": RFC3339}`; DB ไม่ตอบ → `503 {"error":"unavailable"}` | FR-BI-01, AC-BI-01 |
| `GET /api/occupancy/stream` | ไม่มี → `text/event-stream` ส่ง `event: occupancy` `data: {"count": int, "as_of": RFC3339}` ทันทีที่เชื่อมต่อ, ทุกครั้งที่บันทึกสัญญาณสำเร็จ และตอน 00:00 Asia/Bangkok | NFR-PERF-02, AC-BI-04, AC-BI-06, ASM-06 |
| `POST /api/gate/signals` | `{"direction": "in"\|"out"}` → `201 {"count": int, "as_of": RFC3339}`; direction อื่น → `400 {"error":"invalid_direction"}`; DB ไม่ตอบ → `503` | IF-GATE-01 (**ตัวจำลองสัญญาณเท่านั้น ไม่ใช่โปรโตคอลจริงของประตู** Q-01), DOM-CAP-01 |

ทุก endpoint ไม่ต้องยืนยันตัวตน (ASM-07) ดูความเสี่ยงใน D4

### 4.2 Frontend (React Vite)
| หน้าจอ / ส่วน | พฤติกรรม | รองรับ |
|---|---|---|
| `/dashboard` — `OccupancyDashboardPage` (route **ไม่** อยู่ใต้ `RequireAuth`) | โหลดจำนวนจาก `GET /api/occupancy` แล้วเปิด `EventSource("/api/occupancy/stream")` เมื่อมี event ให้แทนค่าตัวเลขบนหน้าจอโดยไม่รีเฟรชหน้า แสดงจำนวนปัจจุบันเป็นข้อมูลหลัก | FR-BI-01, NFR-PERF-02, ASM-07, AC-BI-01, AC-BI-04 (UI-BI-01/02 เป็น Candidate เท่านั้น) |
| `features/occupancy/api.ts` | `fetchOccupancy()` และ `subscribeOccupancy(onCount)` (คืนฟังก์ชันปิด `EventSource`) | FR-BI-01, NFR-PERF-02 |
| `mocks/handlers.ts` | เพิ่ม handler จำลอง `GET /api/occupancy` ตามสัญญา §4.1 | ใช้ทำหน้าจอก่อนต่อ API จริง |

### 4.3 โครงไฟล์ (ให้ `/tasks` ระบุ "ไฟล์ที่แตะ") — ทีมเลือกเอง ไม่ได้มาจาก spec
```
backend/migrations/0003_create_gate_signals.up.sql / .down.sql
backend/internal/occupancy/
  model.go         Direction, ApplySignal (กฎ +1/-1/ไม่ต่ำกว่า 0), StartOfDay (00:00 Asia/Bangkok)
  store.go         RecordSignal (transaction + advisory lock), CurrentCount
  broadcaster.go   กระจายจำนวนใหม่ให้ผู้ฟัง SSE ทุกราย + ตั้งเวลา 00:00
  service.go       ประกอบ store + broadcaster + clock
  handler.go       GET /api/occupancy, GET /api/occupancy/stream, POST /api/gate/signals
  routes.go        Register(gin.IRouter)
  *_test.go        TestAC_BI_xx
backend/cmd/api/main.go                    ประกอบ occupancy handler เข้า router
frontend/src/features/occupancy/
  api.ts, types.ts, OccupancyDashboardPage.tsx, *.test.tsx (ชื่อมี AC_BI_xx)
frontend/src/App.tsx                       เพิ่ม route /dashboard (ไม่ครอบ RequireAuth)
frontend/src/mocks/handlers.ts             handler จำลอง GET /api/occupancy
docs/API.md                                เพิ่มสัญญา API ของ 007 หลัง implement (ขั้นที่ 5 ใน specs/README.md)
```

## 5. ตารางตรวจ Constraints
| Constraint ID | ถูกนำไปใช้ที่ไหนใน plan | สถานะ |
|---|---|---|
| IF-GATE-01 | §4.1 `POST /api/gate/signals` เป็นทางเข้าของตัวจำลองสัญญาณ; §3 `gate_signals` เก็บทุกสัญญาณ; test ทุกตัวส่งสัญญาณผ่านตัวจำลอง | ใช้แล้วผ่านตัวจำลองเท่านั้น — adapter ของประตูจริง **ยังไม่ได้ใช้ เพราะ** ไม่รู้โปรโตคอลและจำนวนประตู (Q-01) |
| DOM-CAP-01 | §3 กฎคำนวณ: `ApplySignal` (+1 / -1 / ไม่ต่ำกว่า 0 ตาม ASM-04), นับใหม่ 00:00 Asia/Bangkok (ASM-03), ประตูบันทึกสองทาง (ASM-01); `occupancy_after >= 0` เป็น CHECK ในฐานข้อมูล | ใช้แล้ว |
| CON-TECH-01 | Go + Gin สำหรับทุก endpoint ใน §4.1 และ package `internal/occupancy` | ใช้แล้ว |
| CON-TECH-02 | React (Vite) หน้า `/dashboard` ใน §4.2 ไม่มีแอปมือถือเนทีฟ | ใช้แล้ว |
| CON-DB-01 (ไม่ผูกกับ 007 ใน SRS) | PostgreSQL ใช้ตาม ASM-02 (ทีมเลือกเอง) | ไม่ใช่ Constraint ของ 007 — บันทึกไว้เพื่อความโปร่งใส |

## 6. แผนทดสอบจาก Acceptance Criteria
| AC ID | ชื่อ test | ทดสอบอย่างไร |
|---|---|---|
| AC-BI-01 | `TestAC_BI_01` (backend), `AC_BI_01` ใน `OccupancyDashboardPage.test.tsx` | Backend (PostgreSQL จริง): ส่งสัญญาณเข้า 5 ออก 2 ผ่าน `POST /api/gate/signals` แล้ว `GET /api/occupancy` ได้ `count` = 3 / Frontend: msw ตอบ `count: 3` แล้วหน้า `/dashboard` แสดง 3 |
| AC-BI-02 | `TestAC_BI_02` | Unit: `ApplySignal(10, in)` = 11; Integration: เตรียมจำนวน 10 แล้วส่ง `in` 1 ครั้ง ได้ 11; subtest ส่ง `in` พร้อมกันหลาย goroutine แล้วจำนวนเท่ากับจำนวนสัญญาณ (`go test -race`) |
| AC-BI-03 | `TestAC_BI_03` | Unit: `ApplySignal(10, out)` = 9; Integration: เตรียมจำนวน 10 แล้วส่ง `out` 1 ครั้ง ได้ 9 |
| AC-BI-04 | `TestAC_BI_04` (backend), `AC_BI_04` ใน `OccupancyDashboardPage.test.tsx` | Backend: `httptest.Server` เปิด `GET /api/occupancy/stream` แล้วส่งสัญญาณ ต้องได้ event ที่มี count ใหม่ภายใน 1 วินาทีนับจากที่ `POST` ตอบกลับ (ASM-06) / Frontend: `EventSource` จำลองยิง event `count: 4` หน้าจอเปลี่ยนจาก 3 เป็น 4 โดยไม่ remount หรือรีเฟรช |
| AC-BI-05 | `TestAC_BI_05` | Unit: `ApplySignal(0, out)` = 0; Integration: จำนวน 0 ส่ง `out` 1 ครั้ง จำนวนยังเป็น 0 และ `gate_signals` มีแถวเพิ่ม 1 แถว (direction `out`) |
| AC-BI-06 | `TestAC_BI_06` | ใช้นาฬิกาจำลอง: บันทึกสัญญาณเข้า 10 ครั้งเวลา 23:59 Asia/Bangkok แล้วเลื่อนนาฬิกาไป 00:00 วันถัดไป `CurrentCount` = 0 และ broadcaster ส่ง event `count: 0` ให้ผู้ฟัง SSE |
| AC-BI-07 | `TestAC_BI_07` | Integration: บันทึก `in` 10 ครั้งผ่าน service ตัวแรก ทิ้ง service/handler นั้น สร้างตัวใหม่ด้วยฐานข้อมูลเดิม (จำลองการรีสตาร์ต) แล้ว `GET /api/occupancy` ได้ 10 |

- คำสั่ง backend: `docker compose up -d postgres` แล้ว `go test -race ./internal/occupancy/...`
- คำสั่ง frontend: `npm test` ใน `frontend/`
- IF-GATE-01 ไม่มี AC; ทุก test ใช้ตัวจำลองสัญญาณ (spec, Q-01)

## 7. ลำดับงาน
1. สร้าง migration `0003_create_gate_signals` (DOM-CAP-01, ASM-02)
2. เขียน `model.go`: `ApplySignal` และ `StartOfDay` (Asia/Bangkok) พร้อม unit test (DOM-CAP-01, AC-BI-02, 03, 05; ASM-03)
3. เขียน `store.go`: `RecordSignal` (transaction + advisory lock) และ `CurrentCount` พร้อม integration test (FR-BI-01, AC-BI-01, 05, 07; ASM-05)
4. เขียน `broadcaster.go` + `service.go`: กระจายจำนวนใหม่ และส่งตอน 00:00 ผ่านนาฬิกาจำลองได้ (NFR-PERF-02, AC-BI-06)
5. เขียน `handler.go` + `routes.go`: `GET /api/occupancy`, `POST /api/gate/signals` (ตัวจำลอง), `GET /api/occupancy/stream` พร้อม handler test (FR-BI-01, IF-GATE-01, AC-BI-01, 04)
6. ประกอบ occupancy handler ใน `backend/cmd/api/main.go` (CON-TECH-01)
7. สร้างหน้า `/dashboard` ด้วย API จำลอง (msw + `EventSource` จำลอง) ไม่ต้องรอ backend (FR-BI-01, NFR-PERF-02, AC-BI-01, 04, ASM-07)
8. ต่อหน้า `/dashboard` กับ API จริง แล้วตรวจด้วยมือบนเบราว์เซอร์: ยิงตัวจำลองสัญญาณแล้วเห็นตัวเลขเปลี่ยนเอง (AC-BI-04)
9. เขียนสัญญา API ของ 007 ใน `docs/API.md` (ขั้นที่ 5 ของ `specs/README.md`)

## 8. สิ่งที่ยังไม่ทำ

### 8.1 Open Questions จาก spec (คัดลอก ไม่ได้ตอบเอง)
| รหัส | คำถาม | สิ่งที่รอ |
|---|---|---|
| Q3 (SRS) | ประตูทางเข้าบันทึกทั้งขาเข้าและขาออกหรือไม่ — ทีมใช้ ASM-01 ชั่วคราว (ยังไม่ได้ถามเจ้าของเรื่องจริง) | plan ใช้กฎ +1/-1 ตาม ASM-01 ถ้าตอบว่าบันทึกแค่ขาเข้า ต้องแก้ DOM-CAP-01, AC-BI-03, 05 |
| Q8 (SRS) | ข้อมูลลงทะเบียนเกี่ยวข้องกับการนับคนเข้ายิมและ Excel เดิมหรือไม่ — อยู่ใน Out of scope | ส่วนที่เกี่ยวข้องกับข้อนี้จะยังไม่สร้างจนกว่าจะได้คำตอบ (ไม่เชื่อมข้อมูลลงทะเบียน) |
| Q10 (SRS) | ปัญหาเครื่องเล่นมีคนใช้เป็นประจำ ทีมต้องการให้ระบบช่วยแก้หรือไม่ — อยู่ใน Out of scope | ส่วนที่เกี่ยวข้องกับข้อนี้จะยังไม่สร้างจนกว่าจะได้คำตอบ (ไม่มีคิวเครื่องเล่น) |
| Q-01 | ประตูส่งสัญญาณมาในรูปแบบใด และมีประตูกี่จุด (IF-GATE-01) — ทีมตอบ: ไม่รู้ ต้องถามฝ่ายอาคารสถานที่ | ส่วนที่เกี่ยวข้องกับข้อนี้จะยังไม่สร้างจนกว่าจะได้คำตอบ (ไม่มี adapter ของประตูจริง ไม่มี `gate_id`; มีเฉพาะ `POST /api/gate/signals` สำหรับตัวจำลอง) |
| Q-02 | รอบเวลาอัปเดตของ NFR-PERF-02 — ทีมใช้ภายใน 1 วินาที (ASM-06) รอยืนยัน | plan ใช้ SSE + test 1 วินาทีตาม ASM-06 ถ้า Product Owner ตอบต่าง ต้องแก้ AC-BI-04 และค่าใน test |
| Q-03 | สัญญาณซ้ำ ผิดลำดับ หรือขาดการเชื่อมต่อ — ทีมใช้รับทุกสัญญาณตามจริง (ASM-05) รอยืนยัน | plan ไม่กรองซ้ำ ไม่เรียงใหม่ ไม่แสดงสถานะประตู |
| Q-04 | ผ่านออกเมื่อจำนวนเป็น 0 — ทีมใช้คงไว้ที่ 0 (ASM-04) รอยืนยัน | plan ใช้ `max(0, n-1)` และเก็บสัญญาณไว้ |
| Q-05 | เก็บเป็นประวัติหรือเฉพาะจำนวน — ทีมใช้เก็บทุกเหตุการณ์ใน PostgreSQL (ASM-02) รอยืนยัน | plan ใช้ตาราง `gate_signals` ไม่มีนโยบายลบข้อมูลเก่า |
| Q-06 | ผู้จัดการล็อกอินอย่างไร และสิทธิ์ดู Dashboard — ทีมใช้ไม่ต้องล็อกอิน (ASM-07) | ส่วนที่เกี่ยวข้องกับข้อนี้จะยังไม่สร้างจนกว่าจะได้คำตอบ (ไม่มี auth บน endpoint และหน้าของ 007) |
| Q-07 | จำนวนปัจจุบันเริ่มนับใหม่เมื่อใด — ทีมใช้ทุกวัน 00:00 (ASM-03) รอยืนยัน | plan ใช้ 00:00 Asia/Bangkok |

ที่เกี่ยวข้องจาก `000-shared` §5: การล็อกอินของเจ้าหน้าที่/ผู้จัดการ (plan ใช้ ASM-07); พฤติกรรมเมื่อ PostgreSQL ไม่ตอบ — spec 007 ไม่กำหนด plan จึงตอบ `503` เท่านั้น ไม่มี fallback (ดู D6)

### 8.2 การออกแบบที่ plan เลือกเอง (ไม่ได้มาจาก spec) ให้ทีมตรวจ
| # | สิ่งที่เลือก | ทางเลือก/ความเสี่ยง |
|---|---|---|
| D1 | ใช้ SSE ส่งจำนวนใหม่ไปยังหน้าเว็บ | ทางเลือก: polling ทุก <1 วินาที (ง่ายแต่ยิง request ถี่) หรือ WebSocket (ต้องเพิ่มไลบรารี และเป็นสองทางโดยไม่จำเป็น) |
| D2 | เก็บ `occupancy_after` ในแต่ละแถวและล็อกด้วย `pg_advisory_xact_lock` | เพราะกฎ "ไม่ต่ำกว่า 0" ทำให้ผลขึ้นกับลำดับ จึงใช้ `SUM` ตรง ๆ ไม่ได้ ข้อเสียคือสัญญาณทุกตัวต้องรอคิวกัน (รับได้เมื่อมีประตูไม่กี่จุด) |
| D3 | เมื่อถึง 00:00 backend ส่ง event `count: 0` เองด้วยตัวตั้งเวลา | ถ้าไม่ส่ง หน้าที่เปิดค้างไว้จะแสดงจำนวนของเมื่อวานจนกว่าจะมีสัญญาณใหม่ ซึ่งขัดกับ AC-BI-06 |
| D4 | `POST /api/gate/signals` ไม่ต้องยืนยันตัวตน | **ใครก็ยิงสัญญาณปลอมได้** รับได้เฉพาะช่วงที่ใช้ตัวจำลอง ต้องแทนด้วย adapter ที่ปลอดภัยเมื่อได้คำตอบ Q-01 |
| D5 | route หน้าเว็บคือ `/dashboard` อยู่นอก `RequireAuth` | ตาม ASM-07 แต่หมายความว่าใครรู้ URL ก็เห็นจำนวนผู้ใช้บริการ |
| D6 | เมื่อ PostgreSQL ไม่ตอบ ให้ตอบ `503` และไม่บันทึกสัญญาณนั้น | spec ไม่กำหนดพฤติกรรมนี้ สัญญาณที่ล้มเหลวจะหายไป (ไม่มี retry หรือ buffer) ต้องให้ Tech Lead ยืนยัน และยังไม่มี AC |
| D7 | ไม่เพิ่ม dependency ใหม่ ใช้ `time/tzdata` ฝัง timezone | ทำให้ binary ใหญ่ขึ้นประมาณ 450 KB |
