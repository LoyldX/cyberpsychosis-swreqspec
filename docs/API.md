# API — Sprint 1 (แบ่งตามฟีเจอร์)

อัปเดต: 11 พฤศจิกายน 2569 | Source: `specs/00N-*/plan.md` และโค้ดใน `backend/` | Constraint: CON-TECH-01 (Go + Gin), CON-TECH-02 (React/Vite), CON-DB-01 (PostgreSQL), CON-CACHE-01 (Redis)

> เอกสารนี้บอก **สัญญา API ที่ทำเสร็จและมี test ยืนยันแล้ว** ของแต่ละฟีเจอร์ ฟีเจอร์ที่ยังไม่ได้ทำระบุไว้ว่า "ยังไม่ได้พัฒนา" ห้ามถือว่ามี endpoint
> ถ้าเอกสารนี้ขัดกับโค้ดหรือ spec ให้ยึด spec (`docs/specs.md`) แล้วตรวจ implementation ก่อนแก้เอกสาร ตัวอย่างและ response ด้านล่างอ้างอิงโค้ดและ test ของแต่ละฟีเจอร์

## สารบัญ
| ฟีเจอร์ | สถานะ |
|---|---|
| [001 auth-session](#001-auth-session) | พัฒนาแล้ว (2 endpoint + middleware) ยกเว้น AC-AUTH-07/08 ยังไม่มี test |
| [004 trainer-roster](#004-trainer-roster) | ยังไม่ได้พัฒนา |
| [005 class-booking](#005-class-booking) | ยังไม่ได้พัฒนา |
| [006 inventory](#006-inventory) | พัฒนาแล้ว (7 endpoints; backend integration tests และ frontend mocked API tests) |
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
**ยังไม่ได้พัฒนา** Spec: `specs/007-occupancy-dashboard/spec.md` — FR-BI-01, NFR-PERF-02, DOM-CAP-01, IF-GATE-01
ยังไม่มี endpoint ที่ตกลงกัน คำถามที่ต้องตอบก่อน: ประตูบันทึกขาเข้า/ขาออกหรือไม่ (Q3), โปรโตคอลของ IF-GATE-01, รอบเวลาอัปเดตของ NFR-PERF-02
