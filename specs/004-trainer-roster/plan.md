# แผนทางเทคนิค: Trainer Roster

Spec: [`spec.md`](spec.md) (Draft v2 ผ่าน /clarify รอบ 2, อัปเดต 12 พฤศจิกายน 2569) — FR-ROST-01, CON-DB-01 | กรอบร่วม: [`specs/000-shared/README.md`](../000-shared/README.md) | ร่าง: 12 พฤศจิกายน 2569 (AI ร่าง รอทีมตรวจ — แทนที่ plan ฉบับ 10 พฤศจิกายน 2569 ที่ยังไม่มี schema เพราะคำถามยังไม่ได้ตอบ)

## 1. สรุปแนวทาง
- เจ้าหน้าที่/ผู้จัดการเปิดหน้าเว็บ `/roster` (ไม่ต้องล็อกอิน ASM-06) เลือกเทรนเนอร์ แล้วเพิ่ม แก้ ยกเลิก หรือลบเวลาเข้างานและเวลาสอนทีละรายการ (FR-ROST-01, ASM-01, ASM-05)
- ตารางงานเป็นรายการซ้ำทุกสัปดาห์: วันในสัปดาห์ + เวลาเริ่ม-จบ ในวันเดียว ตามเวลา Asia/Bangkok (ASM-03)
- ทุกการเปลี่ยนแปลงตรวจกฎใน transaction เดียว: ไม่ทับซ้อน (ต่อกันพอดีได้), เวลาสอนอยู่ในเวลาเข้างาน, ไม่ทำให้คลาสที่มีผู้จองและยังไม่เริ่ม (005) หลุดนอกเวลาเข้างาน ถ้าไม่ผ่านไม่บันทึก (ASM-02, ASM-04, ASM-07, ASM-10)
- ข้อมูลเทรนเนอร์เตรียมไว้ล่วงหน้า (ASM-09); 004 ส่งมอบฟังก์ชัน `Covers` ให้ 005 ใช้ตรวจว่าคลาสอยู่ในเวลาเข้างาน (005 ASM-11)
- ใช้ advisory lock ต่อเทรนเนอร์ key เดียวกับการจองของ 005 การแก้กะกับการจองของเทรนเนอร์คนเดียวกันจึงไม่แทรกกัน (ASM-07)

## 2. เทคโนโลยีที่ใช้
| สิ่งที่เลือก | มาจาก | หมายเหตุ |
|---|---|---|
| Go + Gin | CON-TECH-01 | package ใหม่ `backend/internal/roster/` |
| React (Vite) เว็บอย่างเดียว | CON-TECH-02 | `frontend/src/features/roster/` |
| PostgreSQL | CON-DB-01 | ตาราง `trainers`, `roster_entries` |
| `pgx/v5` + `pgxpool`, `golang-migrate` (`0005_*`) | ทีมเลือกเอง ไม่ได้มาจาก spec (`000-shared` §2) | ต่อจาก 0004 ของ 005 |
| PostgreSQL type `TIME` + `SMALLINT` วันในสัปดาห์ (ISO: จันทร์ = 1 … อาทิตย์ = 7) | ทีมเลือกเอง ไม่ได้มาจาก spec | ตรงกับ ASM-03 (ซ้ำทุกสัปดาห์ ไม่ข้ามเที่ยงคืน) |
| `react-router`, `fetch` + `useEffect`/`useState` | ทีมเลือกเอง ไม่ได้มาจาก spec | ต่างจาก `000-shared` §2 (react-query) ด้วยเหตุผลเดียวกับ 005/007 |
| Go `testing` + testify + `httptest` + PostgreSQL จริง; vitest + Testing Library + msw | ทีมเลือกเอง ไม่ได้มาจาก spec (`000-shared` §2, §3) | |

## 3. โมเดลข้อมูล
migration `0005_create_trainer_roster` (PostgreSQL, CON-DB-01)

| ตาราง | ฟิลด์ | รองรับ | หมายเหตุ |
|---|---|---|---|
| `trainers` | `id BIGINT IDENTITY PK`, `name TEXT NOT NULL` | FR-ROST-01, ASM-09 | เตรียมไว้ล่วงหน้า ไม่มี API เพิ่ม/แก้/ลบ เก็บแค่รหัสและชื่อ |
| `roster_entries` | `id BIGINT IDENTITY PK`, `trainer_id BIGINT NOT NULL REFERENCES trainers(id)`, `kind TEXT NOT NULL CHECK (kind IN ('work','teach'))`, `weekday SMALLINT NOT NULL CHECK (weekday BETWEEN 1 AND 7)`, `start_time TIME NOT NULL`, `end_time TIME NOT NULL`, `cancelled BOOLEAN NOT NULL DEFAULT false` — CHECK `end_time > start_time`, index `(trainer_id, weekday)` | FR-ROST-01, AC-ROST-01..04, ASM-01, ASM-03, ASM-04 | `kind` = เวลาเข้างาน (`work`) / เวลาสอน (`teach`) แยกกัน (ASM-04); CHECK `end_time > start_time` บังคับว่าเริ่ม-จบในวันเดียว (ASM-03); `cancelled` = ยกเลิกแต่ยังเก็บไว้ (ASM-01) ลบ = ลบแถว |

ไม่เพิ่ม foreign key จาก `classes.trainer_id` (005) ไป `trainers` ใน migration นี้ เพราะเป็นตารางของ 005 และข้อมูล/test ของ 005 ยังใช้ `trainer_id` ที่ไม่มีใน `trainers` (005 plan D1) — ดู D1 ข้อ 8.2

## 4. API / หน้าจอ
**ไม่ต้องล็อกอินทุก endpoint** (ASM-06) error เป็น `{"error":"<รหัส>","message":"<ไทย>"}`; รูปแบบรายการ `{id, trainer_id, kind, weekday, start_time:"09:00", end_time:"12:00", cancelled}`

| รายการ | Input / Output หลัก | รองรับ |
|---|---|---|
| `GET /api/trainers` | `200` `[{id, name}]` เรียงตาม id; `503 unavailable` | FR-ROST-01 (เลือกเทรนเนอร์), ASM-09 |
| `GET /api/trainers/:id/roster` | `200` รายการทั้งหมดของเทรนเนอร์ (รวมที่ยกเลิก) เรียงตามวัน เวลาเริ่ม; `404 trainer_not_found`; `503` | AC-ROST-01..04 ("เรียกดูได้") |
| `POST /api/trainers/:id/roster` | Input `{kind, weekday, start_time, end_time}` → `201` รายการ; `400 invalid_request` (รูปแบบผิด, เริ่มไม่ก่อนจบ); `404 trainer_not_found`; `409 overlap` / `teach_outside_work`; `503` | AC-ROST-01, AC-ROST-02, ASM-02, ASM-03, ASM-04, ASM-10 |
| `PATCH /api/roster/:entryId` | Input `{weekday, start_time, end_time}` (เปลี่ยน `kind` ไม่ได้) → `200`; `400`; `404 entry_not_found`; `409 overlap` / `teach_outside_work` / `booked_class_outside_work`; `503` | AC-ROST-03, AC-ROST-04, ASM-05, ASM-07 |
| `POST /api/roster/:entryId/cancel` | → `200` รายการที่ `cancelled: true`; `404`; `409 teach_outside_work` / `booked_class_outside_work`; `503` | ASM-01, ASM-04, ASM-07 |
| `DELETE /api/roster/:entryId` | → `204`; `404`; `409 teach_outside_work` / `booked_class_outside_work`; `503` | ASM-01, ASM-04, ASM-07 |
| `roster.Covers(ctx, trainerID, start, end)` (ฟังก์ชัน Go ไม่ใช่ endpoint) | true ถ้าช่วงเวลาอยู่ในวันเดียว (Asia/Bangkok) และอยู่ในเวลาเข้างานที่ไม่ถูกยกเลิกของวันนั้น (รวมช่วงที่ต่อกัน) | ส่งให้ 005 (005 ASM-11, ASM-13; 004 ASM-10) |
| หน้า `/roster` (React, **ไม่ครอบ** `RequireAuth`) | เลือกเทรนเนอร์ → ตารางรายการแยกเวลาเข้างาน/เวลาสอนตามวัน; ฟอร์มเพิ่ม; ปุ่มแก้/ยกเลิก/ลบต่อรายการ; ข้อความเหตุผลเมื่อ 409; 503 แจ้งลองใหม่และคงค่าในฟอร์ม | FR-ROST-01, CON-TECH-02, ASM-06, ASM-08 (UI-ROST-01 เป็น Candidate) |

กฎตรวจ (ทุกข้อตรวจใน transaction เดียวหลังล็อกเทรนเนอร์ เทียบเฉพาะรายการที่ไม่ถูกยกเลิก):
1. **ทับซ้อน** (ASM-02, ASM-10): รายการ `kind` เดียวกัน เทรนเนอร์เดียวกัน วันเดียวกัน ที่ `start < other.end AND other.start < end` → `409 overlap` (ต่อกันพอดีไม่นับ)
2. **เวลาสอนอยู่ในเวลาเข้างาน** (ASM-04, ASM-10): หลังการเปลี่ยนแปลง ทุกเวลาสอนของวันนั้นต้องอยู่ในเวลาเข้างานของวันเดียวกัน (รวมช่วงที่ต่อกัน) → ไม่ผ่าน `409 teach_outside_work`
3. **คลาสที่มีผู้จองแล้ว** (ASM-07): หลังแก้/ยกเลิก/ลบเวลาเข้างาน ทุกคลาสของเทรนเนอร์นี้ที่มีผู้จองอย่างน้อย 1 คนและยังไม่ถึงเวลาเริ่ม (Q-11) ต้องยัง `Covers` → ไม่ผ่าน `409 booked_class_outside_work`
4. ลำดับตรวจ: รูปแบบ (400) → ไม่พบ (404) → ทับซ้อน → เวลาสอนนอกเวลาเข้างาน → คลาสที่ถูกจอง (D3)

## 5. ตารางตรวจ Constraints
| Constraint ID | ถูกนำไปใช้ที่ไหนใน plan | สถานะ |
|---|---|---|
| CON-DB-01 | ตาราง `trainers`, `roster_entries` (ข้อ 3), transaction + advisory lock ทุกการเปลี่ยนแปลง, test ใช้ PostgreSQL จริง | ใช้แล้ว |
| CON-TECH-01 | Go + Gin ใน `backend/internal/roster/` และ route ใน `cmd/api/main.go` | ใช้แล้ว |
| CON-TECH-02 | หน้า `/roster` ใน React/Vite | ใช้แล้ว |
| DOM | spec ระบุว่าไม่มี | ยังไม่ได้ใช้ เพราะไม่มี DOM ใน spec |
| IF | spec ระบุว่าไม่มี | ยังไม่ได้ใช้ เพราะไม่มี IF ใน spec |

## 6. แผนทดสอบจาก Acceptance Criteria
| AC ID | ชื่อ test | ทดสอบอย่างไร |
|---|---|---|
| AC-ROST-01 | `TestAC_ROST_01_HTTP` | PostgreSQL จริง: เทรนเนอร์ที่ยังไม่มีเวลาเข้างาน → `POST` เวลาเข้างานจันทร์ 09:00–12:00 → `201` แล้ว `GET /api/trainers/:id/roster` เห็นรายการนั้น |
| AC-ROST-01 | `AC_ROST_01` (vitest) | หน้า `/roster` กับ msw: เลือกเทรนเนอร์ กรอกเวลาเข้างาน กดบันทึก → เห็นในตาราง |
| AC-ROST-02 | `TestAC_ROST_02_HTTP` | เทรนเนอร์มีเวลาเข้างานจันทร์ 09:00–12:00 แต่ยังไม่มีเวลาสอน → `POST` เวลาสอน 10:00–11:00 → `201` และเรียกดูได้ (เวลาสอนต้องอยู่ในเวลาเข้างานตาม ASM-04 จึงเตรียมเวลาเข้างานใน Given) |
| AC-ROST-03 | `TestAC_ROST_03_HTTP` | เทรนเนอร์มีตารางอยู่แล้ว → `PATCH` เวลาเข้างานเป็น 13:00–17:00 และเวลาสอน → `GET` ซ้ำเห็นค่าใหม่ |
| AC-ROST-04 | `TestAC_ROST_04_HTTP` | เทรนเนอร์ A และ B มีตารางต่างกัน → แก้/ยกเลิก/ลบของ A → ตารางของ B เหมือนเดิมทุกฟิลด์ |

test เพิ่มที่ไม่ใช่ AC: กฎทับซ้อนและต่อกันพอดี (ASM-10), เวลาสอนนอกเวลาเข้างาน (ASM-04), คลาสที่มีผู้จองและยังไม่เริ่ม/เริ่มแล้ว (ASM-07, Q-11), ยกเลิกแล้วไม่นับเป็นเวลาเข้างาน (ASM-01), รูปแบบเวลาผิด/เริ่มไม่ก่อนจบ (ASM-03), `Covers` ข้ามวันและคร่อมสองกะ, `503` (ASM-08), migration 0005, หน้าเว็บคงค่าในฟอร์มเมื่อ 503 (ASM-08)

## 7. ลำดับงาน
1. migration 0005 `trainers`, `roster_entries` (CON-DB-01)
2. กฎล้วน: แปลง/ตรวจเวลา, ทับซ้อน, อยู่ในช่วงที่ต่อกัน, แปลงช่วงเวลาจริงเป็นวัน+เวลา Asia/Bangkok (ASM-03, ASM-04, ASM-10)
3. store อ่าน + เพิ่มรายการ (ASM-02, ASM-04, ASM-10)
4. store แก้ / ยกเลิก / ลบ พร้อมตรวจคลาสที่ถูกจอง (ASM-01, ASM-05, ASM-07)
5. `Covers` สำหรับ 005
6. Gin handler + route (AC-ROST-01..04)
7. หน้าเว็บ `/roster` ด้วย API จำลอง
8. ประกอบเข้า `main.go` + ข้อมูลเทรนเนอร์ตัวอย่าง + ตรวจด้วยมือ
9. `docs/API.md` หัวข้อ 004

## 8. สิ่งที่ยังไม่ทำ
### 8.1 Open Questions (คัดลอกจาก spec)
| รหัส | คำถาม | สถานะใน spec | ผลต่อ plan |
|---|---|---|---|
| Q-01 | กะซ้ำรายสัปดาห์หรือวันที่ | ทีมตอบแล้ว → ASM-03 (รอยืนยัน) | `weekday` + `TIME` |
| Q-02 | เวลาเข้างาน/เวลาสอนแยกกันไหม | ทีมตอบแล้ว → ASM-04 (รอยืนยัน) | `kind` + กฎข้อ 2 |
| Q-03 | ปรับปรุง = แทนที่หรือแก้รายการ | ทีมตอบแล้ว → ASM-05 (รอยืนยัน) | `PATCH` ทีละรายการ |
| Q-04 | ล็อกอินเจ้าหน้าที่ | ทีมตอบแล้ว → ASM-06 (รอยืนยัน) | ไม่มี auth |
| Q-05 | แก้ตารางทับคลาสที่ถูกจอง | ทีมตอบแล้ว → ASM-07 (รอยืนยัน) | กฎข้อ 3 |
| Q-06 | PostgreSQL ไม่ตอบ | ทีมตอบแล้ว → ASM-08 (รอยืนยัน) | `503` + ฟอร์มคงค่า |
| Q-07 | กะข้ามเที่ยงคืน | ทีมตอบแล้ว → ASM-03 (รอยืนยัน) | CHECK `end_time > start_time` |
| Q-08 | ที่มาของข้อมูลเทรนเนอร์ | ทีมตอบแล้ว → ASM-09 (รอยืนยัน) | ตาราง `trainers` เตรียมไว้ล่วงหน้า |
| Q-09 | ต่อกันพอดีนับทับซ้อนไหม | ทีมตอบแล้ว → ASM-10 (รอยืนยัน) | กฎข้อ 1, 2 และ `Covers` |
| Q-10 | ลบกับยกเลิกต่างกันอย่างไร | ทีมตอบแล้ว → ASM-01 (รอยืนยัน) | `cancelled` |
| Q-11 | ASM-07 นับคลาสที่ผ่านไปแล้วไหม | ทีมตอบแล้ว → ASM-07 (รอยืนยัน) | ตรวจเฉพาะคลาสที่ยังไม่เริ่ม |

ไม่มี Open Question ที่ยังเปิดใน spec 004 รอบนี้

### 8.2 ข้อเสนอการออกแบบที่ทีมต้องยืนยัน (ทีมเลือกเอง ไม่ได้มาจาก spec)
| # | ข้อเสนอ | เหตุผล | ความเสี่ยง |
|---|---|---|---|
| D1 | ยังไม่เพิ่ม FK `classes.trainer_id → trainers(id)` | เป็นตารางของ 005 และ test ของ 005 ใช้ `trainer_id` ที่ไม่มีใน `trainers` | ข้อมูลคลาสอ้างเทรนเนอร์ที่ไม่มีได้ ต้องตัดสินพร้อมทีม 005 |
| D2 | advisory lock `(5005, trainer_id)` key เดียวกับการจองของ 005 | ASM-07: การแก้กะกับการจองของเทรนเนอร์คนเดียวกันต้องไม่แทรกกัน | การแก้ตารางของเทรนเนอร์คนนั้นต้องรอคำขอจองที่ค้างอยู่ |
| D3 | ลำดับตรวจตามกฎข้อ 4 ตอบเหตุผลแรก | spec ไม่กำหนดลำดับ | เห็นเหตุผลเดียวแม้ผิดหลายข้อ |
| D4 | `Covers` อ่านจากตาราง `roster_entries` ด้วย connection แยกจาก transaction ของการจอง | `ShiftChecker` ของ 005 รับแค่ `ctx` | ระหว่างการจองใช้ 2 connection ต่อคำขอ (ยอมรับได้ที่ขนาดยิม) |
| D5 | `GET .../roster` คืนรายการที่ยกเลิกด้วย (มี `cancelled`) | ASM-01: ยกเลิก = ยังเก็บและเห็นในตาราง | — |
| D6 | ยังไม่มีการเลิกยกเลิก (uncancel) | spec/ASM ไม่ได้ขอ | ถ้าต้องการต้องลบแล้วเพิ่มใหม่ |
| D7 | ใช้ `fetch` + `useEffect` ไม่ใช้ react-query | ดูข้อ 2 | ต่างจาก `000-shared` §2 |
| D8 | test ล้าง `roster_entries`, `trainers` (และ `bookings`/`classes` ใน test ของ ASM-07) ในฐานข้อมูล dev | แบบเดียวกับ 005/007 | ข้อมูลตัวอย่างใน dev หายเมื่อรัน test |
| D9 | ข้อมูลเทรนเนอร์ตัวอย่างอยู่ใน `backend/seeds/004_trainers.sql` รันด้วยมือ | ASM-09 | ต้องให้ `trainer_id` ตรงกับข้อมูลคลาสตัวอย่างของ 005 |
