# API — Sprint 1 (แบ่งตามฟีเจอร์)

อัปเดต: 8 พฤศจิกายน 2569 | Source: `specs/00N-*/plan.md` และโค้ดใน `backend/` | Constraint: CON-TECH-01 (Go + Gin), CON-DB-01 (PostgreSQL), CON-CACHE-01 (Redis)

> เอกสารนี้บอก **สัญญา API ที่ทำเสร็จและมี test ยืนยันแล้ว** ของแต่ละฟีเจอร์ ฟีเจอร์ที่ยังไม่ได้ทำระบุไว้ว่า "ยังไม่ได้พัฒนา" ห้ามถือว่ามี endpoint
> ถ้าเอกสารนี้ขัดกับโค้ดหรือ spec ให้ยึด spec (`docs/specs.md`) แล้วแก้เอกสารนี้ ตัวอย่าง `curl` ด้านล่างเขียนตามสัญญาที่ test ตรวจแล้ว และ **ยิงกับ binary จริง (PostgreSQL/Redis จริง) เฉพาะเคสปฏิเสธ** เมื่อ 8 พฤศจิกายน 2569 (login ไม่พบรหัส/JSON เสีย → 401, session ไม่มี header/token ปลอม → 401, route ที่ไม่มี → 404) **เคสสำเร็จ (200) ยังไม่ได้ยิงกับเซิร์ฟเวอร์จริง** เพราะต้องมีสมาชิกในฐานข้อมูลก่อน (Q-03) พิสูจน์ด้วย test เท่านั้น

## สารบัญ
| ฟีเจอร์ | สถานะ |
|---|---|
| [001 auth-session](#001-auth-session) | พัฒนาแล้ว (2 endpoint + middleware) ยกเว้น AC-AUTH-07/08 ยังไม่มี test |
| [004 trainer-roster](#004-trainer-roster) | ยังไม่ได้พัฒนา |
| [005 class-booking](#005-class-booking) | ยังไม่ได้พัฒนา |
| [006 inventory](#006-inventory) | ยังไม่ได้พัฒนา |
| [007 occupancy-dashboard](#007-occupancy-dashboard) | ยังไม่ได้พัฒนา |
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
**ยังไม่ได้พัฒนา** (ยังไม่มี `plan.md` / `tasks.md`) Spec: `specs/004-trainer-roster/spec.md` — FR-ROST-01 (เจ้าหน้าที่/ผู้จัดการจัดตารางเวลาเข้างานและเวลาสอนของเทรนเนอร์)
ยังไม่มี endpoint ที่ตกลงกัน คำถามที่ต้องตอบก่อนออกแบบ API: การล็อกอินและสิทธิ์ของเจ้าหน้าที่/ผู้จัดการ, รูปแบบตารางงาน ดู spec

## 005 class-booking
**ยังไม่ได้พัฒนา** Spec: `specs/005-class-booking/spec.md` — FR-BKG-01, FR-BKG-02, NFR-PERF-01, DOM-BKG-01
จะใช้ `auth.RequireSession` จาก 001 เพื่อรู้ว่าสมาชิกคือใคร และใช้ข้อมูลเวลาปฏิบัติงานจาก 004 ยังไม่มี endpoint ที่ตกลงกัน คำถามที่ต้องตอบก่อน: คลาสเป็นกลุ่มหรือ 1:1, ใครสร้างคลาส, ยกเลิก/ดูรายการจองได้หรือไม่

## 006 inventory
**ยังไม่ได้พัฒนา** Spec: `specs/006-inventory/spec.md` — FR-INV-01..03, NFR-DATA-01, DOM-INV-01, DOM-INV-02
ยังไม่มี endpoint ที่ตกลงกัน คำถามที่ต้องตอบก่อน: ประเภทพัสดุ (Q1), ช่องทางแจ้งเตือนของ FR-INV-03, การล็อกอินของเจ้าหน้าที่

## 007 occupancy-dashboard
**ยังไม่ได้พัฒนา** Spec: `specs/007-occupancy-dashboard/spec.md` — FR-BI-01, NFR-PERF-02, DOM-CAP-01, IF-GATE-01
ยังไม่มี endpoint ที่ตกลงกัน คำถามที่ต้องตอบก่อน: ประตูบันทึกขาเข้า/ขาออกหรือไม่ (Q3), โปรโตคอลของ IF-GATE-01, รอบเวลาอัปเดตของ NFR-PERF-02
