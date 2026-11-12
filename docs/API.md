# API — Sprint 1 (แบ่งตามฟีเจอร์)

อัปเดต: 12 พฤศจิกายน 2569 | Source: `specs/00N-*/plan.md` และโค้ดใน `backend/` | Constraint: CON-TECH-01 (Go + Gin), CON-TECH-02 (React/Vite), CON-DB-01 (PostgreSQL), CON-CACHE-01 (Redis)

> เอกสารนี้บอก **สัญญา API ที่ทำเสร็จและมี test ยืนยันแล้ว** ของแต่ละฟีเจอร์ ฟีเจอร์ที่ยังไม่ได้ทำระบุไว้ว่า "ยังไม่ได้พัฒนา" ห้ามถือว่ามี endpoint
> ถ้าเอกสารนี้ขัดกับโค้ดหรือ spec ให้ยึด spec (`docs/specs.md`) แล้วตรวจ implementation ก่อนแก้เอกสาร ตัวอย่างและ response ด้านล่างอ้างอิงโค้ดและ test ของแต่ละฟีเจอร์

## สารบัญ
| ฟีเจอร์ | สถานะ |
|---|---|
| [001 auth-session](#001-auth-session) | พัฒนาแล้ว (2 endpoint + middleware) ยกเว้น AC-AUTH-07/08 ยังไม่มี test |
| [004 trainer-roster](#004-trainer-roster) | พัฒนาแล้ว (6 endpoint ไม่ต้องล็อกอิน) — test กับ PostgreSQL จริงผ่าน, ยังไม่ได้ตรวจด้วยมือ |
| [005 class-booking](#005-class-booking) | พัฒนาแล้ว (3 endpoint หลัง `RequireSession`, ตรวจกะจากตารางงานของ 004) — test กับ PostgreSQL จริงผ่าน, ยังไม่ได้ตรวจด้วยมือ |
| [006 inventory](#006-inventory) | พัฒนาแล้ว (7 endpoints; backend integration tests และ frontend mocked API tests) |
| [007 occupancy-dashboard](#007-occupancy-dashboard) | พัฒนาแล้ว (3 endpoint รวม SSE) — test กับ PostgreSQL จริงผ่านแล้ว |
| 002 notification, 003 line-bot, 008 network-segmentation | ไม่อยู่ใน sprint นี้ |

## ธรรมเนียมร่วมของทุก endpoint
- **Base path:** `/api/...` พอร์ตเริ่มต้นของเซิร์ฟเวอร์คือ `:8080` (ตั้งด้วย env `HTTP_ADDR`)
- **รูปแบบข้อมูล:** JSON ทั้งขาเข้าและขาออก (`Content-Type: application/json`)
- **การยืนยันตัวตน:** header `Authorization: Bearer <refresh_token>` (token ที่ได้จาก `POST /api/auth/login`) ฟีเจอร์ที่ต้องรู้ว่าผู้เรียกคือสมาชิกคนไหนให้ห่อ route ด้วย middleware `auth.RequireSession` (ดูหัวข้อ 001)
- **รูปแบบ error:** `{"error": "<รหัส>"}` (บางกรณีมี `"message"` ภาษาไทยเพิ่ม)

| HTTP | `error` | ความหมาย |
|---|---|---|
| 401 | `invalid_credentials` | เข้าสู่ระบบไม่สำเร็จ (ใช้ข้อความเดียวกันทุกกรณี) |
| 401 | `unauthorized` | ไม่มี/ผิดรูปแบบ/ไม่พบ/หมดอายุของ token |
| 503 | `service_unavailable` | Redis หรือ PostgreSQL มีปัญหา ให้ลองใหม่ (มี `message` ภาษาไทย) |

---

## 001 auth-session
Spec: `specs/001-auth-session/spec.md` (Draft v2) | FR-AUTH-01, NFR-REL-01, CON-CACHE-01 | ผู้ใช้: สมาชิก

**สมมติฐานของทีมที่ API นี้อิงอยู่ (ASM ใน spec, ยังไม่ได้ยืนยันกับเจ้าของเรื่องจริง):** เข้าสู่ระบบด้วยรหัสนักศึกษา+รหัสผ่านที่ระบบเก็บเอง (ASM-01), Refresh Token อายุ 7 วันนับจากเข้าสู่ระบบ ไม่ต่ออายุ (ASM-02), ไม่มี logout (ASM-03), หลายอุปกรณ์พร้อมกันได้ (ASM-04), ระบบล่มให้ปฏิเสธ (ASM-05), ไม่จำกัดจำนวนครั้งที่เข้าสู่ระบบผิด (ASM-06)

### POST /api/auth/login
เข้าสู่ระบบและรับ Refresh Token (FR-AUTH-01; AC-AUTH-01, 05, 06)

**Request**
```json
{ "student_id": "6607xxxxx", "password": "••••••••" }
```

**Response**
| HTTP | Body | เมื่อ |
|---|---|---|
| 200 | `{"refresh_token": "<สตริงสุ่ม>", "expires_in": 604800}` | รหัสนักศึกษาและรหัสผ่านถูกต้อง (`expires_in` เป็นวินาที = 7 วัน) |
| 401 | `{"error":"invalid_credentials"}` | ไม่พบรหัสนักศึกษา **หรือ** รหัสผ่านผิด **หรือ** body เสีย/ฟิลด์ว่าง/หาย — body เหมือนกันทุกตัวอักษร ไม่บอกว่าผิดข้อใด |
| 503 | `{"error":"service_unavailable","message":"ระบบไม่พร้อมให้บริการ กรุณาลองใหม่อีกครั้ง"}` | Redis หรือ PostgreSQL มีปัญหา |

**หมายเหตุ**
- ทุกครั้งที่เข้าสู่ระบบสำเร็จจะได้ token ใหม่เสมอ และไม่ลบ token เดิม (หลายอุปกรณ์ได้ ASM-04)
- ฝั่งเซิร์ฟเวอร์เก็บเฉพาะค่า SHA-256 ของ token ใน Redis (key `session:<sha256-hex>`) และไม่เก็บ token ดิบที่ใด
- เมื่อไม่พบรหัสนักศึกษา เซิร์ฟเวอร์ยังเทียบ bcrypt กับ hash หลอกเพื่อให้เวลาตอบใกล้เคียงกับกรณีรหัสผิด

```bash
curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"student_id":"6607xxxxx","password":"********"}'
```

### GET /api/auth/session
ตรวจว่า token ยังใช้ได้และดูว่าเป็นสมาชิกคนไหน = "ขอใช้งานต่อ" (FR-AUTH-01; AC-AUTH-02, 03, 04, 07, 09, 10)

**Request header:** `Authorization: Bearer <refresh_token>` (ไม่ใส่ body)

**Response**
| HTTP | Body | เมื่อ |
|---|---|---|
| 200 | `{"member": {"id": 1, "student_id": "6607xxxxx"}}` | token ยังไม่หมดอายุ |
| 401 | `{"error":"unauthorized"}` | ไม่มี header, scheme ไม่ใช่ `Bearer`, token ว่าง/มีหลายตัว, token ไม่เคยออก/ถูกแก้ไข, หรือหมดอายุ — body เหมือนกันทุกกรณี |
| 503 | `{"error":"service_unavailable","message":"ระบบไม่พร้อมให้บริการ กรุณาลองใหม่อีกครั้ง"}` | Redis มีปัญหา |

**หมายเหตุ**
- scheme `Bearer` ไม่สนตัวพิมพ์เล็กใหญ่ ต้องมี token ตัวเดียวพอดี
- การเรียก endpoint นี้ **ไม่ต่ออายุ** token (อายุนับจากเข้าสู่ระบบเท่านั้น)
- ตามการออกแบบ D2 ใน `plan.md` endpoint นี้อ่านข้อมูลสมาชิกจาก Redis **ไม่ถาม PostgreSQL** จึงยังตอบ 200 ได้แม้ PostgreSQL ล่ม ซึ่งอาจขัดกับตัวอักษรของ ASM-05 ทีมยังไม่ได้ตัดสิน
- ฝั่งเว็บ: ถ้าตอบ 401 ให้ลบ token ออกจาก `localStorage` และไปหน้าเข้าสู่ระบบ ถ้าตอบ 503 หรือเครือข่ายล้ม **ห้ามลบ token** ให้แสดงปุ่มลองใหม่

```bash
curl -s http://localhost:8080/api/auth/session \
  -H 'Authorization: Bearer <refresh_token>'
```

### Middleware `RequireSession` (สำหรับฟีเจอร์อื่นในโค้ด Go)
ใช้ใน `backend/internal/auth/middleware.go` (FR-AUTH-01; AC-AUTH-02)

```go
// ห่อ route ที่ต้องเข้าสู่ระบบ
r.GET("/api/booking", auth.RequireSession(sessions), handler)

// ใน handler ดึงสมาชิกที่ยืนยันแล้ว
data, ok := auth.MemberFromContext(c) // data.MemberID, data.StudentID
```
- ไม่มี/ผิดรูปแบบ/ไม่พบ/หมดอายุ → ตอบ 401 `{"error":"unauthorized"}` และหยุด chain; Redis มีปัญหา → 503 และหยุด chain
- ตอนนี้ `sessions` ต่อกับ `*auth.SessionStore` ใน `backend/cmd/api/main.go` ฟีเจอร์ 005 ต้องใช้ `data.MemberID` เป็น member ที่การจองอ้างถึง

### สถานะการทดสอบ (ตาม `specs/001-auth-session/tasks.md`)
| AC | Test | สถานะ |
|---|---|---|
| AC-AUTH-01, 05 | `TestAC_AUTH_01`, `TestAC_AUTH_05` | ผ่าน (PostgreSQL/Redis จริง) |
| AC-AUTH-02, 04 | `TestAC_AUTH_02`, `TestAC_AUTH_04` | ผ่าน |
| AC-AUTH-03 | `TestAC_AUTH_03` | ผ่าน (พิสูจน์กลไกหมดอายุด้วย TTL 1 วินาที; ค่า 7 วันพิสูจน์ที่ `TestAC_AUTH_02`) |
| AC-AUTH-06 | `TestAC_AUTH_06` และ test เสริม | ผ่าน |
| AC-AUTH-09 | `TestAC_AUTH_09` | ผ่าน |
| AC-AUTH-10 | `AC_AUTH_10: resume after reload` (vitest) | ผ่านระดับจำลอง; **ต้องตรวจด้วยมือบนเบราว์เซอร์จริง (T-16)** |
| AC-AUTH-07, 08 | — | **ยังไม่มี test** (ข้ามงาน blackhole/timeout ตามคำสั่งทีม; รอ Q-04) |

### ข้อจำกัดที่รู้อยู่
- **ไม่มี timeout ที่เรากำหนดเอง** (Q-04) ถ้า Redis/PostgreSQL ค้างโดยไม่ตอบ คำขอจะรอตามค่าเริ่มต้นของไลบรารี ไม่ใช่ 503 ภายในเวลาที่กำหนด ส่วน error ที่ไลบรารีส่งกลับมาจะถูกตอบ 503
- **ไม่มีการสร้างบัญชีสมาชิกใน API** (Q-03) ตาราง `members` ต้องมีข้อมูลอยู่ก่อน
- **ไม่มีขั้นตอนรัน migration ในแอป** ตาราง `members` สร้างจาก `backend/migrations/0001_create_members.up.sql` (test ใช้ `golang-migrate` เป็นไลบรารีรันให้) การติดตั้งใหม่ต้องรันไฟล์นี้เอง
- **ไม่มี rate limit** การเข้าสู่ระบบผิดซ้ำ (ASM-06)
- **ไม่มี logout/เพิกถอน token** (ASM-03)

### วิธีรัน backend
```bash
docker compose up -d postgres redis
cd backend
DATABASE_URL='postgres://gym_user:gym_password_dev@localhost:5432/gym_management?sslmode=disable' \
REDIS_ADDR='localhost:6379' \
go run ./cmd/api
```
ค่า `DATABASE_URL` ข้างบนเป็นค่า dev จาก `docker-compose.yml` ไม่ใช่ค่าสำหรับ production; ถ้าไม่ตั้ง `DATABASE_URL` หรือ `REDIS_ADDR` โปรแกรมไม่เริ่มและบอกสาเหตุ
ถ้าไม่ตั้ง `GIN_MODE=release` Gin จะรันใน debug mode (แสดงรายการ route และคำเตือนใน log) ตั้งเมื่อใช้งานจริง

---

## 004 trainer-roster
Spec: [`specs/004-trainer-roster/spec.md`](../specs/004-trainer-roster/spec.md) (Draft v2 ผ่าน /clarify รอบ 2) — FR-ROST-01, CON-DB-01 | Backend: Go/Gin (CON-TECH-01) + PostgreSQL (CON-DB-01) | Web: React/Vite หน้า `/roster` (CON-TECH-02) | โค้ด: `backend/internal/roster/`

> **สถานะการทดสอบ (12 พฤศจิกายน 2569):** test ทั้งหมดของ `backend/internal/roster` ผ่านกับ PostgreSQL จริง (`go test -race`) รวม `TestAC_ROST_01_HTTP`..`TestAC_ROST_04_HTTP`; ต่อเข้า `main.go` แล้วและตรวจกับเซิร์ฟเวอร์จริง; หน้าเว็บทดสอบด้วย API จำลอง (msw); **ยังไม่ได้ตรวจด้วยมือบนเบราว์เซอร์** (ดู `docs/test_report/004-trainer-roster/manual_test.md`)

### ธรรมเนียม Roster API
- **ไม่ต้องล็อกอินทุก endpoint** (ASM-06) — ใครเข้าถึงเซิร์ฟเวอร์ได้ก็แก้ตารางงานได้
- **ตารางงานซ้ำทุกสัปดาห์** (ASM-03): `weekday` ISO (จันทร์ = 1 … อาทิตย์ = 7), `start_time`/`end_time` เป็น `"HH:MM"` เวลา Asia/Bangkok ในวันเดียว (เริ่มก่อนจบ ไม่ข้ามเที่ยงคืน)
- **`kind`:** `work` = เวลาเข้างาน, `teach` = เวลาสอน (แยกกัน ASM-04)
- **ยกเลิก vs ลบ** (ASM-01): ยกเลิก = ยังเก็บไว้ (`cancelled: true`) แต่ไม่นับเป็นเวลาเข้างาน/เวลาสอน, ลบ = หายจากระบบ
- **ข้อมูลเทรนเนอร์เตรียมไว้ล่วงหน้า** ไม่มี API เพิ่ม/แก้/ลบเทรนเนอร์ (ASM-09) ตัวอย่างอยู่ใน `backend/seeds/004_trainers.sql`
- **รูปแบบ error:** `{"error": "<รหัส>", "message": "<ข้อความภาษาไทย>"}`

| Method | Path | รองรับ |
|---|---|---|
| GET | `/api/trainers` | FR-ROST-01 (เลือกเทรนเนอร์) |
| GET | `/api/trainers/{id}/roster` | AC-ROST-01..04 (เรียกดู) |
| POST | `/api/trainers/{id}/roster` | AC-ROST-01, AC-ROST-02 |
| PATCH | `/api/roster/{entryId}` | AC-ROST-03, AC-ROST-04 |
| POST | `/api/roster/{entryId}/cancel` | ASM-01 |
| DELETE | `/api/roster/{entryId}` | ASM-01 |

### รูปแบบ Trainer และ Entry
```json
{"id": 1001, "name": "ครูเอ (ตัวอย่าง)"}
```
```json
{"id": 61, "trainer_id": 1001, "kind": "work", "weekday": 1, "start_time": "09:00", "end_time": "12:00", "cancelled": false}
```

### GET /api/trainers
- **200:** JSON array ของ Trainer เรียงตาม `id` (ไม่มีได้ `[]`) | **503** `unavailable`

### GET /api/trainers/{id}/roster
- **200:** JSON array ของ Entry ทั้งหมดของเทรนเนอร์ **รวมที่ยกเลิก** เรียงตามวันและเวลาเริ่ม (plan D5)
- **400** `invalid_request` (`{id}` ไม่ใช่จำนวนเต็มบวก) | **404** `trainer_not_found` | **503** `unavailable`

### POST /api/trainers/{id}/roster
Request: `{"kind": "work", "weekday": 1, "start_time": "09:00", "end_time": "12:00"}` → **201** Entry

### PATCH /api/roster/{entryId}
Request: `{"weekday": 1, "start_time": "09:00", "end_time": "13:00"}` (เปลี่ยน `kind` ไม่ได้) → **200** Entry

### POST /api/roster/{entryId}/cancel และ DELETE /api/roster/{entryId}
- cancel → **200** Entry ที่ `cancelled: true` (ยกเลิกซ้ำได้ผลเดิม) | delete → **204** ไม่มี body

### Roster API errors
| HTTP | `error` | เมื่อไร |
|---|---|---|
| 400 | `invalid_request` | body ไม่ใช่ JSON, ไม่มี `weekday`, `kind` ไม่ใช่ work/teach, `weekday` นอก 1–7, เวลาไม่ใช่ `HH:MM`, เวลาเริ่มไม่ก่อนเวลาจบ, หรือรหัสใน path ผิดรูป |
| 404 | `trainer_not_found` / `entry_not_found` | ไม่มีเทรนเนอร์หรือรายการนั้น |
| 409 | `overlap` | ทับกับรายการประเภทเดียวกันของเทรนเนอร์คนเดียวกันในวันเดียวกัน (ต่อกันพอดีไม่นับ ASM-02, ASM-10) |
| 409 | `teach_outside_work` | หลังการเปลี่ยนแปลงมีเวลาสอนอยู่นอกเวลาเข้างาน (เวลาเข้างานที่ต่อกันนับรวมกัน ASM-04, ASM-10) |
| 409 | `booked_class_outside_work` | แก้/ยกเลิก/ลบเวลาเข้างานแล้วคลาสของ 005 ที่มีผู้จองและยังไม่เริ่ม ซึ่งเคยอยู่ในกะ จะอยู่นอกกะ (ASM-07, Q-11) |
| 503 | `unavailable` | ฐานข้อมูลไม่ตอบ ไม่บันทึกอะไรเลย (ASM-08) |

- ตรวจตามลำดับ: รูปแบบ → ไม่พบ → ทับซ้อน → เวลาสอนนอกเวลาเข้างาน → คลาสที่ถูกจอง (plan D3)
- การเปลี่ยนแปลงของเทรนเนอร์คนเดียวกันใช้ advisory lock key เดียวกับการจองของ 005 จึงไม่แทรกกัน (plan D2)
- ฟังก์ชัน Go `(*roster.Store).Covers(ctx, trainerID, start, end)` (ไม่ใช่ endpoint) ให้ 005 ตรวจว่าคลาสอยู่ในเวลาเข้างาน

## 005 class-booking
Spec: [`specs/005-class-booking/spec.md`](../specs/005-class-booking/spec.md) (Draft v2 ผ่าน /clarify รอบ 2) — FR-BKG-01, FR-BKG-02, NFR-PERF-01, DOM-BKG-01, CON-DB-01 | Backend: Go/Gin (CON-TECH-01) + PostgreSQL (CON-DB-01) | Web: React/Vite หน้า `/classes` (CON-TECH-02) | โค้ด: `backend/internal/booking/`

> **ต่อเข้าเซิร์ฟเวอร์แล้ว (12 พฤศจิกายน 2569, T-11):** route อยู่หลัง `auth.RequireSession` ใน `backend/cmd/api/main.go` และตรวจกะจากเวลาเข้างานจริงของ 004 ผ่าน `booking.NewRosterShifts(roster.Store)` — ตรวจกับเซิร์ฟเวอร์จริงแล้ว (ล็อกอินจริง → จองในกะ 201, นอกกะ 409)
> **สถานะการทดสอบ:** test ทั้งหมดของ `backend/internal/booking` ผ่านกับ PostgreSQL จริง (`go test -race`) รวม `TestAC_BKG_01_HTTP`, `TestAC_BKG_02_HTTP`, `TestAC_BKG_03_HTTP`, `TestAC_BKG_04_HTTP`, `TestAC_BKG_05` (20 รอบ ตอบภายใน 1 วินาที) — test ระดับ API ส่วนใหญ่ใช้กะจำลอง และ `TestAC_BKG_03_Roster`/`_04_Roster` ใช้เวลาเข้างานจริงของ 004; หน้าเว็บทดสอบด้วย API จำลอง (msw); ยังไม่ได้ตรวจด้วยมือ (ดู `docs/test_report/005-class-booking/manual_test.md`)

### ธรรมเนียม Booking API
- **ทุก endpoint ต้องล็อกอิน:** ห่อด้วย `auth.RequireSession` ของ 001 ไม่มี/ผิด token ได้ `401 {"error":"unauthorized"}` และ Redis ล่มได้ `503 {"error":"service_unavailable"}` จาก middleware
- **ผู้จองคือสมาชิกใน session** (`members.id` ของ 001, ASM-06) ไม่รับ member id จาก body
- **รูปแบบ error ของ 005:** `{"error": "<รหัส>", "message": "<ข้อความภาษาไทย>"}`
- **ข้อมูลคลาสเตรียมไว้ล่วงหน้าในฐานข้อมูล** ไม่มี API สร้าง/แก้คลาส (ASM-04) และไม่มี API ยกเลิกการจองหรือดูรายการจองทั้งหมด (Out of scope, Q-04)
- **เวลา:** ISO 8601 (RFC 3339) มี offset เขตเวลาเสมอ ให้ฝั่งเว็บแปลงเป็นเวลา Asia/Bangkok เอง

| Method | Path | รองรับ |
|---|---|---|
| GET | `/api/classes` | FR-BKG-01 (เลือกคลาส) |
| POST | `/api/bookings` | FR-BKG-01, FR-BKG-02, DOM-BKG-01, NFR-PERF-01 |
| GET | `/api/bookings/{id}` | FR-BKG-01 (AC-BKG-01 "เรียกดูการจองนั้นได้") |

### รูปแบบ Class และ Booking
```json
{"id": 11, "trainer_id": 1, "starts_at": "2026-11-13T10:00:00+07:00", "ends_at": "2026-11-13T11:00:00+07:00", "capacity": 10}
```
```json
{"id": 101, "class_id": 11, "member_id": 1, "created_at": "2026-11-13T08:00:00+07:00"}
```
`trainer_id` ยังไม่ผูก foreign key กับข้อมูลเทรนเนอร์ของ 004 (plan D1)

### GET /api/classes
- **200:** JSON array ของ Class ที่ยังไม่ถึงเวลาเริ่ม เรียงตาม `starts_at` (ไม่มีคลาสได้ `[]`) ไม่แสดงจำนวนที่นั่งเหลือ (plan D4)
- **503:** `unavailable`

### POST /api/bookings
Request: `{"class_id": 11}`

| HTTP | `error` | เมื่อไร |
|---|---|---|
| 201 | — | จองสำเร็จ คืน Booking |
| 400 | `invalid_request` | body ไม่ใช่ JSON, ไม่มี `class_id`, หรือ `class_id` ไม่ใช่จำนวนเต็มบวก |
| 404 | `class_not_found` | ไม่มีคลาสนี้ |
| 409 | `class_started` | ถึงเวลาเริ่มคลาสแล้ว (ASM-05, Q-12) |
| 409 | `already_booked` | สมาชิกจองคลาสนี้ไว้แล้ว (ASM-05) |
| 409 | `class_full` | ที่นั่งเต็ม (ASM-03, ASM-05) |
| 409 | `trainer_off_shift` | คลาสไม่อยู่ในเวลาเข้างานของเทรนเนอร์ (ตารางงานของ 004) ทั้งหมด ขอบพอดีนับว่าอยู่ในกะ เวลาเข้างานที่ต่อกันนับรวม หรือเทรนเนอร์ไม่มีกะวันนั้น (DOM-BKG-01, ASM-08, ASM-11, ASM-13; 004 ASM-10) |
| 409 | `trainer_busy` | เวลาทับกับคลาสอื่นของเทรนเนอร์คนเดียวกันที่มีผู้จองแล้ว ต่อกันพอดีนับว่าทับ (DOM-BKG-01, ASM-09, ASM-10, ASM-12) |
| 503 | `unavailable` | ฐานข้อมูลหรือแหล่งข้อมูลกะไม่ตอบ ไม่บันทึกอะไรเลย ไม่ retry (ASM-07) |

- ถ้าผิดหลายข้อ ตอบเหตุผลแรกตามลำดับในตาราง (`class_not_found` → `class_started` → `already_booked` → `class_full` → `trainer_off_shift` → `trainer_busy`, plan D3)
- คำขอของเทรนเนอร์คนเดียวกันถูกเรียงทีละรายการใน transaction (advisory lock, plan D2) จึงไม่เกิดการจองซ้อนเมื่อยิงพร้อมกัน (AC-BKG-05)

ตัวอย่าง:
```sh
curl -i -X POST http://localhost:8080/api/bookings \
  -H 'Authorization: Bearer <refresh_token>' -H 'Content-Type: application/json' \
  -d '{"class_id": 11}'
# HTTP/1.1 409 Conflict
# {"error":"trainer_busy","message":"จองไม่สำเร็จ เพราะเทรนเนอร์ไม่ว่าง มีคิวสอนคลาสอื่นในช่วงเวลานี้"}
```

### GET /api/bookings/{id}
- **200:** Booking ถ้าเป็นของสมาชิกใน session
- **404 `booking_not_found`:** ไม่พบ หรือเป็นของสมาชิกคนอื่น (ตอบเหมือนกัน plan D5)
- **400 `invalid_request`:** `{id}` ไม่ใช่จำนวนเต็มบวก
- **503 `unavailable`**

### ยังไม่ได้พัฒนา
- ตรวจด้วยมือบนเบราว์เซอร์ (T-12) — ข้อมูลคลาสตัวอย่างอยู่ใน `backend/seeds/005_classes.sql` แล้ว
- ยกเลิกการจอง / ดูรายการจองของตนเอง — Q-04 ยังเปิด ต้องเสนอ FR ใหม่ก่อน

## 006 inventory
Spec: [`specs/006-inventory/spec.md`](../specs/006-inventory/spec.md) — FR-INV-01..03, NFR-DATA-01, DOM-INV-01, DOM-INV-02 | Backend: Go/Gin + PostgreSQL (CON-TECH-01, CON-DB-01); Web client: React/Vite (CON-TECH-02)

Inventory API จัดการรายการอุปกรณ์ยืม-คืนและวัสดุสิ้นเปลือง (FR-INV-01), บันทึกการเบิก/ยืม/คืน/รับเข้า (FR-INV-02) และคืนรายการที่ถึง Reorder Point (FR-INV-03). API นี้ไม่มีการบังคับ authentication/authorization; เรื่องดังกล่าวอยู่นอก scope ของ spec.

### ธรรมเนียม Inventory API
- Base path: `/api/inventory`; ตัวอย่างด้านล่างใช้ `http://localhost:8080`
- Request/response ใช้ JSON ยกเว้น `DELETE` สำเร็จตอบ `204 No Content`
- `item_type` ใน Web UI ใช้ค่า `equipment` หรือ `consumable`; API เก็บค่าที่รับมาโดยไม่มี enum validation ใน handler ปัจจุบัน
- `reorder_point` เป็น nullable และไม่บังคับ (DOM-INV-02); รายการที่สร้างใหม่เริ่มด้วย `quantity_on_hand: 0`
- การเบิก/ยืมลดสต็อก ส่วนคืน/รับเข้าเพิ่มสต็อก; การปรับยอดและบันทึก movement อยู่ใน transaction เดียว ยอดเบิก/ยืมไม่ติดลบ (NFR-DATA-01)

### Endpoints ที่ implement แล้ว
| Method | Path | รายละเอียด | สำเร็จ |
|---|---|---|---|
| `POST` | `/api/inventory` | สร้างรายการ (FR-INV-01) | `201` |
| `GET` | `/api/inventory` | อ่านรายการที่มองเห็น; ค้นหาชื่อบางส่วนแบบไม่สนตัวพิมพ์ด้วย `?name=` (FR-INV-01) | `200` |
| `GET` | `/api/inventory/alerts` | อ่านรายการที่มี Reorder Point และยอดคงเหลือไม่เกินจุด (FR-INV-03, DOM-INV-02) | `200` |
| `GET` | `/api/inventory/{id}` | อ่านรายละเอียดรายการ (FR-INV-01) | `200` |
| `PATCH` | `/api/inventory/{id}` | แก้ไข `name`, `item_type` และ/หรือ `reorder_point` (FR-INV-01) | `200` |
| `DELETE` | `/api/inventory/{id}` | ซ่อนรายการจาก catalog โดยเก็บประวัติ movement ไว้ (FR-INV-01, FR-INV-02) | `204` |
| `POST` | `/api/inventory/{id}/movements` | บันทึกการเบิก ยืม คืน หรือรับเข้า (FR-INV-02) | `201` |

`GET /api/inventory/{id}/movements` ปรากฏในแผน แต่ยังไม่มีการลงทะเบียน route ใน implementation ปัจจุบัน จึงยังไม่ใช่ endpoint ที่ใช้งานได้และไม่รวมในรายการข้างต้น.

### รูปแบบ Inventory Item
ใช้เป็น response ของการอ่าน/สร้าง/แก้ไขรายการ และเป็นสมาชิกของ response แบบ array:
```json
{
  "id": 12,
  "name": "วิทยุสื่อสาร",
  "item_type": "equipment",
  "quantity_on_hand": 5,
  "reorder_point": 2
}
```
`reorder_point` จะเป็นตัวเลขหรือ `null`.

### POST /api/inventory
สร้างรายการใหม่ (FR-INV-01, DOM-INV-02). `name` และ `item_type` จำเป็น; `reorder_point` ไม่บังคับ. ไม่รับยอดเริ่มต้นจาก request.

**Request**
```json
{
  "name": "วิทยุสื่อสาร",
  "item_type": "equipment",
  "reorder_point": 2
}
```

**Response `201 Created`**
```json
{
  "id": 12,
  "name": "วิทยุสื่อสาร",
  "item_type": "equipment",
  "quantity_on_hand": 0,
  "reorder_point": 2
}
```

### GET /api/inventory
คืน array ของรายการที่ยังไม่ถูกซ่อน เรียงตาม `id`. หากไม่พบรายการจะได้ `[]`.

| Query | ความหมาย |
|---|---|
| `name` | ค้นหาชื่อแบบ contains และไม่สนตัวพิมพ์เล็ก/ใหญ่; ไม่ส่งหรือส่งค่าว่างเพื่ออ่านทั้งหมด |

ตัวอย่าง: `GET /api/inventory?name=วิทยุ`

**Response `200 OK`**
```json
[
  {
    "id": 12,
    "name": "วิทยุสื่อสาร",
    "item_type": "equipment",
    "quantity_on_hand": 5,
    "reorder_point": 2
  }
]
```

### GET /api/inventory/alerts
คืน array เฉพาะรายการที่ยังมองเห็น มี `reorder_point` และ `quantity_on_hand <= reorder_point` (FR-INV-03, DOM-INV-02). เมื่อไม่มีรายการตรงเงื่อนไข คืน `[]`. รูปแบบแต่ละรายการเหมือน [Inventory Item](#รูปแบบ-inventory-item).

### GET /api/inventory/{id}
คืนรายละเอียดและยอดปัจจุบันของรายการที่ยังมองเห็น โดย body เป็น Inventory Item.

### PATCH /api/inventory/{id}
ส่งเฉพาะฟิลด์ที่ต้องการแก้ไขได้: `name`, `item_type`, `reorder_point`. ต้องมีอย่างน้อยหนึ่งฟิลด์; ส่ง `"reorder_point": null` เพื่อล้าง Reorder Point.

**Request**
```json
{ "name": "วิทยุสื่อสารรุ่นใหม่", "reorder_point": null }
```

**Response `200 OK`**: Inventory Item ที่แก้ไขแล้ว.

### DELETE /api/inventory/{id}
ซ่อนรายการจาก catalog โดยไม่ลบข้อมูล movement ที่เกี่ยวข้อง.

**Response `204 No Content`**: ไม่มี response body.

### POST /api/inventory/{id}/movements
บันทึก movement และปรับจำนวนคงเหลือใน transaction เดียว (FR-INV-02, DOM-INV-01, NFR-DATA-01).

**Request fields**
| Field | Type | ข้อกำหนด |
|---|---|---|
| `movement_type` | string | `issue`, `borrow`, `return`, หรือ `receive` |
| `quantity` | integer | ต้องมากกว่า 0 |
| `responsible_person` | string | ต้องไม่ว่างสำหรับ `issue`, `borrow`, `return` |
| `purpose` | string | ต้องไม่ว่างสำหรับ `issue`, `borrow`, `return` |
| `due_date` | string (`YYYY-MM-DD`) | จำเป็นเฉพาะ `borrow`; ห้ามส่งกับ movement ชนิดอื่น |

`receive` ไม่บังคับ `responsible_person` หรือ `purpose`. ตัวอย่างการยืม:
```json
{
  "movement_type": "borrow",
  "quantity": 1,
  "responsible_person": "สมชาย",
  "purpose": "ใช้งานในหน่วยงาน",
  "due_date": "2026-11-20"
}
```

**Response `201 Created`**
```json
{
  "movement": {
    "id": 34,
    "inventory_item_id": 12,
    "movement_type": "borrow",
    "quantity": 1,
    "responsible_person": "สมชาย",
    "purpose": "ใช้งานในหน่วยงาน",
    "due_date": "2026-11-20",
    "created_at": "2026-11-11T08:00:00Z"
  },
  "quantity_on_hand": 4
}
```

### Inventory API errors
Error response ใช้ JSON `{"error":"<code>"}`.

| HTTP | `error` | เมื่อ |
|---|---|---|
| `400` | `invalid_request` | JSON/ฟิลด์ไม่ถูกต้อง, ID ไม่ใช่จำนวนเต็มบวก, movement ไม่รองรับ, จำนวนไม่เป็นบวก, ข้อมูลบังคับขาด, หรือรูปแบบวันไม่ถูกต้อง |
| `404` | `not_found` | ไม่พบรายการที่ระบุ หรือรายการถูกซ่อนแล้ว |
| `409` | `insufficient_stock` | เบิก/ยืมมากกว่ายอดคงเหลือ |
| `500` | `internal_error` | เกิดข้อผิดพลาดภายใน/ฐานข้อมูล; movement ที่ล้มเหลวไม่ทำให้ยอดหรือประวัติบันทึกเพียงบางส่วน |

ตัวอย่าง:
```bash
curl -i 'http://localhost:8080/api/inventory?name=วิทยุ'
curl -i -X POST 'http://localhost:8080/api/inventory/12/movements' \
  -H 'Content-Type: application/json' \
  -d '{"movement_type":"issue","quantity":1,"responsible_person":"สมชาย","purpose":"ใช้งานในหน่วยงาน"}'
```

### ยังไม่ได้พัฒนา
- อ่านประวัติ movement ผ่าน `GET /api/inventory/{id}/movements` (มีใน plan แต่ไม่มี route ในโค้ด)
- Web UI แจ้งเตือนแบบ push/notification; ปัจจุบันหน้าเว็บอ่านรายการแจ้งเตือนจาก endpoint `/alerts` เท่านั้น (FR-INV-03)

## 007 occupancy-dashboard
Spec: [`specs/007-occupancy-dashboard/spec.md`](../specs/007-occupancy-dashboard/spec.md) (Draft v2) — FR-BI-01, NFR-PERF-02, DOM-CAP-01, IF-GATE-01 | Backend: Go/Gin (CON-TECH-01) + PostgreSQL (ASM-02 ทีมเลือกเอง; SRS ไม่ผูก 007 กับ CON-DB-01) | Web: React/Vite หน้า `/dashboard` (CON-TECH-02) | โค้ด: `backend/internal/occupancy/`

> **สถานะการทดสอบ (12 พฤศจิกายน 2569):** test ทั้งหมดของ `backend/internal/occupancy` (24 test รวม `TestAC_BI_01_HTTP`, `_02_`, `_03_`, `_05_`, `_07_`, `TestAC_BI_04`, `TestAC_BI_06`, store และ migration) **ผ่านกับ PostgreSQL จริง** (`go test -race`) — ทดสอบบน PostgreSQL 18 เพราะเครื่องที่รันไม่มี Docker (`docker-compose.yml` ใช้ 16) และทีมตรวจด้วยมือบนเบราว์เซอร์ (M-01..M-04) ผ่านแล้ว ดู `docs/test_report/007-occupancy-dashboard/manual_test.md`

### ธรรมเนียม Occupancy API
- **ไม่ต้องยืนยันตัวตนทุก endpoint** (ASM-07) — ใครเข้าถึงเซิร์ฟเวอร์ได้ก็อ่านจำนวนและยิงสัญญาณจำลองได้ (plan D4, D5)
- **จำนวนปัจจุบัน** = เข้า +1, ออก -1, ไม่ต่ำกว่า 0 (DOM-CAP-01, ASM-04) นับเฉพาะสัญญาณตั้งแต่ 00:00 Asia/Bangkok ของวันนี้ (ASM-03) รับทุกสัญญาณตามลำดับที่ได้รับ ไม่กรองตัวซ้ำ (ASM-05)
- รูปแบบ **Snapshot** ที่ทุก endpoint คืน:
```json
{ "count": 3, "as_of": "2026-11-12T12:00:00.123456+07:00" }
```
`as_of` คือเวลาที่เซิร์ฟเวอร์คำนวณจำนวนนี้ (RFC3339)

### Endpoints ที่ implement แล้ว
| Method | Path | รายละเอียด | สำเร็จ |
|---|---|---|---|
| `GET` | `/api/occupancy` | อ่านจำนวนปัจจุบัน (FR-BI-01) | `200` Snapshot |
| `GET` | `/api/occupancy/stream` | Server-Sent Events ของจำนวนปัจจุบัน (NFR-PERF-02) | `200 text/event-stream` |
| `POST` | `/api/gate/signals` | **ตัวจำลองสัญญาณประตูเท่านั้น** ไม่ใช่โปรโตคอลจริงของประตู (IF-GATE-01, Q-01) | `201` Snapshot |

### GET /api/occupancy
```sh
curl http://localhost:8080/api/occupancy
```
**Response `200 OK`** — Snapshot เช่น `{"count":3,"as_of":"..."}` (ถ้าวันนี้ยังไม่มีสัญญาณ `count` เป็น `0`)

### GET /api/occupancy/stream
เปิดค้างไว้เพื่อรับจำนวนใหม่โดยไม่ต้องรีเฟรช (NFR-PERF-02, ASM-06: ภายใน 1 วินาทีหลังระบบรับสัญญาณ) ฝั่งเว็บใช้ `new EventSource("/api/occupancy/stream")`
```sh
curl -N http://localhost:8080/api/occupancy/stream
```
**Response `200`** headers: `Content-Type: text/event-stream`, `Cache-Control: no-cache`; เซิร์ฟเวอร์ส่ง event ชื่อ `occupancy` ซึ่ง `data` เป็น Snapshot
```
event:occupancy
data:{"count":3,"as_of":"2026-11-12T12:00:00+07:00"}

```
ส่ง event เมื่อ: (1) เชื่อมต่อสำเร็จ (จำนวนปัจจุบัน) (2) ทุกครั้งที่บันทึกสัญญาณสำเร็จ (3) ตอน 00:00 Asia/Bangkok (จำนวนของวันใหม่ ปกติคือ `0`) — ผู้ฟังที่อ่านไม่ทันจะได้ค่าล่าสุด ไม่ได้ทุกค่า ไม่มี keep-alive ping

### POST /api/gate/signals
บันทึกสัญญาณผ่านประตู 1 ครั้ง ณ เวลาที่เซิร์ฟเวอร์ได้รับ แล้วกระจายจำนวนใหม่ให้ทุก stream (DOM-CAP-01, IF-GATE-01) — ใช้กับตัวจำลองและการทดสอบเท่านั้นจนกว่าจะรู้โปรโตคอลจริง (Q-01)

**Request**
```json
{ "direction": "in" }
```
`direction` ต้องเป็น `"in"` หรือ `"out"` เท่านั้น (ตรงตามตัวพิมพ์)

**Response `201 Created`** — Snapshot หลังสัญญาณนี้ เช่น `{"count":4,"as_of":"..."}`; สัญญาณ `out` ตอนจำนวนเป็น 0 ได้ `201` และ `count: 0` และยังถูกบันทึกไว้ (ASM-04)

```sh
curl -X POST -H 'Content-Type: application/json' -d '{"direction":"in"}' http://localhost:8080/api/gate/signals
```

### Occupancy API errors
| HTTP | `error` | เมื่อไร |
|---|---|---|
| `400` | `invalid_direction` | `POST /api/gate/signals` ที่ body ไม่ใช่ JSON หรือ `direction` ไม่ใช่ `in`/`out` — ไม่บันทึกสัญญาณ |
| `503` | `unavailable` | ฐานข้อมูลใช้ไม่ได้ ทั้ง 3 endpoint (stream ตอบ 503 แทนการเปิด stream) — สัญญาณที่ล้มเหลว **ไม่ถูกบันทึกและไม่มี retry** (plan D6 รอ Tech Lead ยืนยัน) |

### ยังไม่ได้พัฒนา
- Adapter รับสัญญาณจากประตูจริง และการแยกประตูหลายจุด (T-12 รอ Q-01)
- การยืนยันตัวตน/สิทธิ์ของผู้จัดการ (Out of scope, ASM-07)
