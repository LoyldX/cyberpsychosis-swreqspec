# Plan: Auth Session Persistence (001)
Spec: `specs/001-auth-session/spec.md` (Draft v2) | Spec ID: FR-AUTH-01, NFR-REL-01, CON-CACHE-01, CON-DB-01 | ร่างโดย AI ด้วย /plan | Status: Draft v1 | Owner: _(รอมอบหมาย)_ | Reviewed by: _(รอผู้ตรวจ)_ | อัปเดต: 8 พฤศจิกายน 2569
ร่วมกับ: `specs/000-shared/README.md` (Constraints ร่วมและเทคโนโลยีที่ทีมเลือก)

## 1. สรุปแนวทาง
- ฟีเจอร์นี้ให้สมาชิกเข้าสู่ระบบด้วยรหัสนักศึกษา+รหัสผ่าน (ASM-01) แล้วใช้งานต่อได้ภายใน 7 วันโดยไม่เข้าสู่ระบบใหม่ (FR-AUTH-01, ASM-02)
- ผู้ใช้คือสมาชิก (นักศึกษา) ผ่านเว็บ React (Vite) (CON-TECH-02) และเป็นฐานให้ฟีเจอร์อื่น (เช่น 005) รู้ว่าผู้เรียกคือสมาชิกคนไหน
- Backend เป็น Go + Gin (CON-TECH-01) เก็บสมาชิกและ password hash ใน PostgreSQL (CON-DB-01) เก็บ Session ใน Redis พร้อมอายุ 7 วันที่ไม่ต่ออายุ (CON-CACHE-01)
- Refresh Token เป็นสตริงสุ่มที่ไม่มีความหมายในตัว (opaque) เก็บเฉพาะค่า hash ใน Redis หลายอุปกรณ์ได้ไม่จำกัด (ASM-04) ไม่มี logout (ASM-03)
- เมื่อ Redis หรือ PostgreSQL ไม่ตอบ ระบบปฏิเสธและแจ้งให้ลองใหม่ (ASM-05) ฝั่งเว็บเก็บ token ไว้ข้ามการปิดเบราว์เซอร์ (ASM-07)

## 2. เทคโนโลยีที่ใช้

| สิ่งที่เลือก | มาจาก | หมายเหตุ |
|---|---|---|
| Go + Gin framework (Backend API) | CON-TECH-01 | ค่าเริ่มต้นของ `/plan` (Python FastAPI) ไม่ใช้ เพราะขัดกับ CON-TECH-01 |
| React (Vite) เว็บอย่างเดียว | CON-TECH-02 | ไม่มีแอปมือถือเนทีฟ |
| PostgreSQL 16 | CON-DB-01 | image จาก `docker-compose.yml` |
| Redis 7 | CON-CACHE-01 | image จาก `docker-compose.yml` |
| `pgx/v5` (`pgxpool`) เขียน SQL ตรง | ทีมเลือกเอง ไม่ได้มาจาก spec (`000-shared` §2) | |
| `golang-migrate` | ทีมเลือกเอง ไม่ได้มาจาก spec (`000-shared` §2) | migration ใน `backend/migrations/` |
| `go-redis/v9` | ทีมเลือกเอง ไม่ได้มาจาก spec (`000-shared` §2) | |
| `golang.org/x/crypto/bcrypt` | ทีมเลือกเอง ไม่ได้มาจาก spec | **ใหม่ ยังไม่อยู่ใน `000-shared` §2 ขอทีมยืนยัน** ใช้ hash รหัสผ่าน (ASM-01) |
| `crypto/rand` + SHA-256 (stdlib) | ทีมเลือกเอง ไม่ได้มาจาก spec | สร้าง token และ hash ก่อนเก็บใน Redis |
| `react-router`, `fetch` + `@tanstack/react-query` | ทีมเลือกเอง ไม่ได้มาจาก spec (`000-shared` §2) | |
| `testing` + `testify`, `httptest` | ทีมเลือกเอง ไม่ได้มาจาก spec (`000-shared` §3) | integration test ใช้ PostgreSQL/Redis จริง ห้าม mock |
| `vitest` + `@testing-library/react` + `msw` | ทีมเลือกเอง ไม่ได้มาจาก spec (`000-shared` §2) | |

## 3. โมเดลข้อมูล

**PostgreSQL — ตาราง `members`** (CON-DB-01)

| ฟิลด์ | ชนิด | หมายเหตุ | รองรับ |
|---|---|---|---|
| `id` | BIGINT identity, PK | ฟีเจอร์อื่น (เช่น 005) อ้างถึง | FR-AUTH-01 |
| `student_id` | TEXT, UNIQUE, NOT NULL | รหัสนักศึกษา | FR-AUTH-01, ASM-01 |
| `password_hash` | TEXT, NOT NULL | ผลจาก bcrypt ห้ามเก็บรหัสผ่านจริง | FR-AUTH-01, ASM-01 |
| `created_at` | TIMESTAMPTZ, default now() | | — |

- ไม่มีฟิลด์ชื่อ/อีเมล/เบอร์ เพราะ spec 001 ไม่ต้องใช้ (ฟีเจอร์อื่นเพิ่มได้ผ่าน migration ของตน)
- **ไม่มีตารางเก็บ Refresh Token ใน PostgreSQL** เพราะ CON-CACHE-01 ให้ Redis บริหาร Session และ ASM-03 ไม่มีการเพิกถอน
- ไม่มีช่องทางสร้างสมาชิกในโค้ด production (Q-03 ยังไม่ตอบ) test สร้างข้อมูลจำลองผ่าน helper ใน `_test.go` เท่านั้น

**Redis — Session** (CON-CACHE-01)

| รายการ | ค่า | รองรับ |
|---|---|---|
| key | `session:<sha256-hex ของ Refresh Token>` | AC-AUTH-05 |
| value | JSON `{ "member_id": <id>, "student_id": "<รหัสนักศึกษา>" }` | FR-AUTH-01 |
| TTL | 7 วัน (604800 วินาที) ตั้งตอนเข้าสู่ระบบ **ไม่ต่ออายุเมื่อใช้** | ASM-02 |
| หลายอุปกรณ์ | การเข้าสู่ระบบแต่ละครั้งสร้าง key ใหม่ ไม่ลบของเดิม | ASM-04 |

## 4. API / หน้าจอ

### 4.1 Backend (Gin)
| Method path | Input → Output หลัก | รองรับ |
|---|---|---|
| `POST /api/auth/login` | `{student_id, password}` → 200 `{refresh_token, expires_in}`; 401 `{error: "invalid_credentials"}` ข้อความเดียวกันทั้ง "ไม่พบรหัส" และ "รหัสผ่านผิด"; 503 `{error: "service_unavailable"}` พร้อมข้อความให้ลองใหม่ | FR-AUTH-01; AC-AUTH-01, 06, 08; ASM-01, 05, 06 |
| `GET /api/auth/session` | header `Authorization: Bearer <refresh_token>` → 200 `{member: {id, student_id}}`; 401 เมื่อหมดอายุ/ไม่ถูกต้อง; 503 เมื่อ Redis ไม่ตอบ | FR-AUTH-01; AC-AUTH-02, 03, 04, 07, 09; ASM-02, 05 |
| middleware `RequireSession` (ใช้ซ้ำโดยฟีเจอร์อื่น) | ตรวจ token แบบเดียวกับ `GET /api/auth/session` แล้วใส่ `member_id` ใน context | FR-AUTH-01; AC-AUTH-02 ("เรียกฟังก์ชันที่ต้องเข้าสู่ระบบ") |

- ไม่มี endpoint ออกจากระบบ ลงทะเบียน ตั้งรหัสผ่าน หรือต่ออายุ (ASM-03, Q-03 และอยู่นอก scope)
- ไม่ทำ rate limit การเข้าสู่ระบบ (ASM-06)

### 4.2 Frontend (React Vite)
| หน้าจอ/ส่วนประกอบ | หน้าที่ | รองรับ |
|---|---|---|
| `/login` — `LoginPage` | ฟอร์มรหัสนักศึกษา+รหัสผ่าน; สำเร็จแล้วเก็บ `refresh_token` ใน `localStorage` ไปหน้าหลัก; ผิดแสดงข้อความทั่วไป | AC-AUTH-01, 06 |
| `AuthProvider` | ตอนเปิดเว็บ อ่าน token แล้วเรียก `GET /api/auth/session`: 200 → เข้าสู่ระบบอยู่; 401 → ลบ token ไป `/login`; 503/เครือข่ายล้ม → แสดง "ลองใหม่" **ไม่ลบ token** | AC-AUTH-02, 03, 07, 10; ASM-05, 07 |
| `RequireAuth` (route guard) | ไม่มีสถานะเข้าสู่ระบบ → ไป `/login` | AC-AUTH-10 |
| `/` — `HomePage` | แสดงรหัสนักศึกษาของสมาชิกที่ยืนยันแล้ว (เพื่อให้เห็นว่าใช้งานต่อได้) ไม่มีปุ่มออกจากระบบ (ASM-03) | AC-AUTH-10; UI-AUTH-01 (Candidate) |

### 4.3 โครงไฟล์ (ให้ `/tasks` ระบุ "ไฟล์ที่แตะ") — ทีมเลือกเอง ไม่ได้มาจาก spec
```
docker-compose.yml                      (แก้: เอาการ mount backend/db/schema.sql ที่ไม่มีแล้วออก)
backend/go.mod
backend/cmd/api/main.go
backend/internal/config/config.go
backend/internal/app/router.go
backend/internal/auth/{handler.go, service.go, session_store.go, member_store.go, middleware.go}
backend/internal/auth/*_test.go
backend/internal/testutil/{fixtures.go, blackhole.go}
backend/migrations/0001_create_members.{up,down}.sql
frontend/package.json, vite.config.ts
frontend/src/features/auth/{api.ts, AuthProvider.tsx, RequireAuth.tsx, LoginPage.tsx, HomePage.tsx}
frontend/src/features/auth/*.test.tsx
frontend/src/mocks/{handlers.ts, server.ts}   (msw จำลองตามสัญญาใน §4.1)
```

## 5. ตารางตรวจ Constraints

| Constraint ID | ถูกนำไปใช้ที่ไหนใน plan | สถานะ |
|---|---|---|
| CON-CACHE-01 | §3 Redis เก็บ Session (key/TTL/หลายอุปกรณ์), §4.1 ทุก endpoint ตรวจ Session จาก Redis, AC-AUTH-05, 07 | ใช้แล้ว |
| CON-DB-01 | §2 PostgreSQL, §3 ตาราง `members`, AC-AUTH-08 | ใช้แล้ว (ค่า timeout ของการเชื่อมต่อรอ Q-04) |
| CON-TECH-01 | §2 Go + Gin, §4.1 และ §4.3 โครง `backend/` | ใช้แล้ว |
| CON-TECH-02 | §2 React (Vite) เว็บอย่างเดียว, §4.2 และ §4.3 โครง `frontend/` ไม่มีแอปเนทีฟ | ใช้แล้ว |
| DOM-* | spec 001 ไม่มี DOM | ไม่มีใน spec นี้ |
| IF-* | spec 001 ไม่มี IF | ไม่มีใน spec นี้ |

## 6. แผนทดสอบจาก Acceptance Criteria

ชื่อ test ตามธรรมเนียมใน `000-shared` §4 (Go: `TestAC_AUTH_nn`, React: ชื่อมี `AC_AUTH_nn`) integration test ต้อง `docker compose up -d postgres redis` และรัน `go test -race ./...`

| AC ID | ชื่อ test | ทดสอบอย่างไร |
|---|---|---|
| AC-AUTH-01 | `TestAC_AUTH_01` | สร้างสมาชิกจำลองด้วย helper (bcrypt hash) → `POST /api/auth/login` ด้วยรหัสถูกต้อง → 200 และมี `refresh_token` ไม่ว่าง (PostgreSQL/Redis จริง) |
| AC-AUTH-02 | `TestAC_AUTH_02` | เข้าสู่ระบบ → `GET /api/auth/session` ด้วย token → 200 ได้ `member.id` ตรงกับที่สร้าง; เรียก route ทดสอบที่ห่อด้วย `RequireSession` สำเร็จโดยไม่ส่งรหัสผ่าน; ตรวจ `TTL` ของ key ใน Redis อยู่ในช่วง 7 วัน (604800 วินาที ลบเศษไม่เกิน 1 นาที) |
| AC-AUTH-03 | `TestAC_AUTH_03` | สร้าง service ด้วย TTL สั้นใน test (เช่น 1 วินาที) → เข้าสู่ระบบ → รอเกิน TTL → `GET /api/auth/session` ได้ 401 (หมายเหตุ: รอ 7 วันจริงไม่ได้ จึงพิสูจน์คู่กับการตรวจ TTL = 7 วันใน AC-AUTH-02) |
| AC-AUTH-04 | `TestAC_AUTH_04` | token สุ่มที่ไม่เคยออก และ token จริงที่แก้ 1 ตัวอักษร → ทั้งคู่ได้ 401 |
| AC-AUTH-05 | `TestAC_AUTH_05` | หลังเข้าสู่ระบบ คำนวณ SHA-256 ของ token → ตรวจ key `session:<hash>` มีอยู่ใน Redis จริง และ `member_id` ตรงกัน |
| AC-AUTH-06 | `TestAC_AUTH_06` | ทดสอบ 2 กรณี (ไม่มีรหัสนักศึกษา / รหัสผ่านผิด) → 401 ทั้งคู่ ไม่มีฟิลด์ `refresh_token` และ body ของสองกรณีเหมือนกันทุกตัวอักษร |
| AC-AUTH-07 | `TestAC_AUTH_07` | ใช้ Redis client ชี้ไปที่ตัวจำลอง (blackhole: รับ TCP แต่ไม่ตอบ) พร้อม timeout ใน test config → `GET /api/auth/session` ด้วย token ที่ถูกต้อง → 503 พร้อมข้อความให้ลองใหม่ ภายในเวลา timeout |
| AC-AUTH-08 | `TestAC_AUTH_08` | ใช้ pgxpool ชี้ไปที่ blackhole + timeout ใน test config → `POST /api/auth/login` → 503 พร้อมข้อความให้ลองใหม่ |
| AC-AUTH-09 | `TestAC_AUTH_09` | เข้าสู่ระบบสองครั้ง (อุปกรณ์ A, B) ได้ token ต่างกัน → ทั้งสอง token เรียก `GET /api/auth/session` ได้ 200 |
| AC-AUTH-10 | `AC_AUTH_10: resume after reload` (vitest + msw) | ใส่ token ไว้ใน `localStorage` แล้ว mount แอปใหม่ (จำลองเปิดเว็บใหม่) โดย msw ตอบ 200 → เห็นรหัสนักศึกษาบน `HomePage` ไม่เห็นฟอร์มเข้าสู่ระบบ + **ตรวจด้วยมือบนเบราว์เซอร์จริง** (ปิดทั้งโปรแกรมแล้วเปิดใหม่) เพราะ vitest ปิดเบราว์เซอร์จริงไม่ได้ |

NFR-REL-01 ไม่มี AC (รอ Q-01) จึงไม่มี test

## 7. ลำดับงาน
1. **โครง backend และ docker-compose** — สร้าง `backend/` (Go module, Gin app แยกจาก server, config) แก้ `docker-compose.yml` ให้ไม่ mount ไฟล์ที่ไม่มี (CON-TECH-01, CON-DB-01, CON-CACHE-01)
2. **Migration และ member store** — `members` + ฟังก์ชันค้นหาตามรหัสนักศึกษา (FR-AUTH-01; รองรับ AC-AUTH-01, 06, 08)
3. **Session store (Redis)** — สร้าง token, hash, เก็บ/อ่านพร้อม TTL 7 วัน (CON-CACHE-01, ASM-02; AC-AUTH-02, 03, 04, 05, 09)
4. **Login** — ตรวจรหัสผ่านด้วย bcrypt, ข้อความผิดเหมือนกัน, ออก token (FR-AUTH-01; AC-AUTH-01, 06)
5. **Session endpoint และ `RequireSession`** (FR-AUTH-01; AC-AUTH-02, 03, 04, 09)
6. **จัดการระบบล่ม** — แปลง timeout/การเชื่อมต่อล้มของ Redis/PostgreSQL เป็น 503 (ASM-05; AC-AUTH-07, 08) **ค่า timeout รอ Q-04**
7. **Go test ครบ AC-AUTH-01..09** พร้อม helper สร้างสมาชิกจำลองและตัวจำลอง blackhole
8. **โครง frontend** — Vite + msw จำลอง API ตามสัญญา §4.1 (CON-TECH-02)
9. **LoginPage, AuthProvider, RequireAuth, HomePage** + test `AC_AUTH_10` (FR-AUTH-01, ASM-07)
10. **ต่อ frontend กับ API จริง** และตรวจด้วยมือว่าปิด-เปิดเบราว์เซอร์แล้วยังอยู่ในระบบ (AC-AUTH-10)

## 8. สิ่งที่ยังไม่ทำ

### 8.1 Open Questions จาก spec (คัดลอก ไม่ได้ตอบเอง)
| รหัส | คำถาม | สิ่งที่รอ |
|---|---|---|
| Q2 (SRS) | กลุ่มสมาชิกที่โดน Auto-logout และ Session ต้องคงอยู่นานเท่าใด — ทีมยืนยัน ASM-02 แล้ว (ยังไม่ได้ถามเจ้าของเรื่องจริง) | plan ใช้ 7 วันตาม ASM-02 ถ้าเจ้าของเรื่องตอบต่าง ต้องแก้ TTL และ AC-AUTH-02, 03 |
| Q7 (SRS) | ผู้ใช้บริการมีเฉพาะนักศึกษาหรือไม่ — ทีมยืนยัน ASM-01 แล้ว (ยังไม่ได้ถามเจ้าของเรื่องจริง) | plan ใช้รหัสนักศึกษา+รหัสผ่านตาม ASM-01 ถ้าตอบต่าง ต้องแก้วิธีเข้าสู่ระบบ |
| Q-01 | อัตราการหลุดโดยไม่ตั้งใจสูงสุดที่ยอมรับได้ (NFR-REL-01) เท่าใด นับช่วงเวลาใด — ทีมตอบ: ไม่รู้ ต้องถาม Product Owner | ส่วนที่เกี่ยวข้องกับข้อนี้จะยังไม่สร้างจนกว่าจะได้คำตอบ (ไม่มี test ของ NFR-REL-01) |
| Q-02 | **ปิดแล้วสำหรับ sprint นี้:** ไม่มี logout (ASM-03) ถ้าต้องการในอนาคตให้เสนอ FR ใหม่เข้า SRS ก่อน | ไม่มีส่วนที่รอ (plan ไม่มี endpoint ออกจากระบบ) |
| Q-03 | บัญชีสมาชิกและรหัสผ่านเริ่มต้นมาจากไหน — ทีมตอบ: ไม่รู้ ต้องถาม | ส่วนที่เกี่ยวข้องกับข้อนี้จะยังไม่สร้างจนกว่าจะได้คำตอบ (ไม่มีการสร้าง/นำเข้าสมาชิกในโค้ด production ใช้ข้อมูลจำลองใน test เท่านั้น) |
| Q-04 | กี่วินาทีที่ถือว่า Redis หรือ PostgreSQL "ไม่ตอบสนอง" | ส่วนที่เกี่ยวข้องกับข้อนี้จะยังไม่สร้างจนกว่าจะได้คำตอบ (ค่า timeout เริ่มต้นใน config; test ตั้งค่าเองในไฟล์ test) |

ที่เกี่ยวข้องจาก `000-shared` §5: ยืนยันว่า "Member Mobile App" ส่งมอบเป็นเว็บ responsive (plan ใช้ ASM-07); พฤติกรรมเมื่อ Redis/PostgreSQL ไม่ตอบ (plan ใช้ ASM-05)

### 8.2 การออกแบบที่ plan เลือกเอง (ไม่ได้มาจาก spec) ให้ทีมตรวจ
| # | สิ่งที่เลือก | ทางเลือก/ความเสี่ยง |
|---|---|---|
| D1 | ใช้ token ตัวเดียว (Refresh Token แบบ opaque) เป็นทั้งตัวต่ออายุและตัวยืนยันตัวตนของ endpoint อื่น เพราะ spec กำหนดแค่ Refresh Token | ถ้าต้องการ access token อายุสั้นแยกต่างหาก ต้องมีข้อกำหนดเพิ่ม (spec ไม่มี) |
| D2 | `GET /api/auth/session` ไม่เรียก PostgreSQL (ข้อมูลสมาชิกอยู่ใน Redis) ทำให้ **ถ้า PostgreSQL ล่ม การขอใช้งานต่อยังสำเร็จ** ขณะที่ ASM-05 เขียนว่า "Redis หรือ PostgreSQL ไม่ตอบสนอง ระบบปฏิเสธการเข้าสู่ระบบ/ขอใช้งานต่อ" และ AC ครอบคลุมเฉพาะ Redis ล่มตอนขอใช้งานต่อ (AC-AUTH-07) กับ PostgreSQL ล่มตอนเข้าสู่ระบบ (AC-AUTH-08) | ทีมต้องตัดสินว่า ASM-05 ต้องให้ขอใช้งานต่อล้มเมื่อ PostgreSQL ล่มด้วยหรือไม่ (ถ้าใช่ ต้องให้ session endpoint ตรวจ PostgreSQL ทุกครั้ง และเพิ่ม AC) |
| D3 | เก็บ token ใน `localStorage` (อยู่ข้ามการปิดเบราว์เซอร์ได้ ตาม ASM-07) | เสี่ยงถูกอ่านเมื่อมีช่องโหว่ XSS ทางเลือกคือ httpOnly cookie แต่ต้องจัดการ CORS/CSRF เพิ่ม |
| D4 | ไม่มี rate limit การเข้าสู่ระบบ ตาม ASM-06 | เดารหัสผ่านได้ไม่จำกัดครั้ง |
| D5 | `bcrypt` ยังไม่อยู่ใน `000-shared` §2 | ขอทีมยืนยันแล้วเพิ่มใน `000-shared` |
