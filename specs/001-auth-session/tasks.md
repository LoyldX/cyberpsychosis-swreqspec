# Tasks: Auth Session Persistence (001)
Feature: Auth Session Persistence | Spec ID: FR-AUTH-01, NFR-REL-01, CON-CACHE-01, CON-DB-01 | อ้างอิง: `specs/001-auth-session/plan.md`, `specs/001-auth-session/spec.md` (Draft v2) | ร่างโดย AI ด้วย /tasks | อัปเดต: 8 พฤศจิกายน 2569

**สรุป:** ทั้งหมด 16 task, รอ Open Question 2 task (T-11, T-12 รอ Q-04), อีก 14 task สถานะ "พร้อมทำ" (ต้องทำตามลำดับใน "ต้องทำหลัง")
**ข้อควรทราบ:** plan §8.2 มี 5 เรื่องที่ทีมยังไม่ได้ตรวจ โดยเฉพาะ D2 (session endpoint ไม่เรียก PostgreSQL) ซึ่งถ้าทีมตัดสินต่างจาก plan จะกระทบ T-08, T-12 และต้องเพิ่ม AC

ทุก task แตะเฉพาะไฟล์ที่ระบุ รวมถึง `go.mod`/`go.sum` และ `package.json`/`package-lock.json` เมื่อเพิ่มไลบรารี ชื่อ test ตามธรรมเนียม `specs/000-shared/README.md` §4 (Go: `TestAC_AUTH_nn`, React: ชื่อมี `AC_AUTH_nn`)

---

## ฝั่ง Backend (Go + Gin)

### T-01 สร้างโครง backend (Go module, Gin app แยกจาก server, config พื้นฐาน)
- รองรับ: CON-TECH-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-06, T-08
- ไฟล์ที่แตะ: `backend/go.mod`, `backend/go.sum`, `backend/cmd/api/main.go`, `backend/internal/config/config.go`, `backend/internal/app/router.go`, `backend/internal/app/router_test.go`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: `go build ./...` และ `go test ./...` ผ่าน (test ยืนยันว่าสร้าง router ได้) โดย config อ่านเฉพาะ address ของเซิร์ฟเวอร์, `DATABASE_URL`, `REDIS_ADDR` ยังไม่มีค่า timeout (รอ Q-04 ที่ T-11)
- สถานะ: เสร็จ รอทีมตรวจ (coordinator รัน build/vet/test ซ้ำแล้ว 8 พฤศจิกายน 2569; ไม่มี test ของ config เพราะ config_test.go ไม่อยู่ในรายการไฟล์)

### T-02 แก้ docker-compose.yml ให้ PostgreSQL และ Redis ขึ้นได้
- รองรับ: CON-DB-01, CON-CACHE-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-03, T-05
- ไฟล์ที่แตะ: `docker-compose.yml`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: เอาบรรทัด mount `./backend/db/schema.sql` ที่ไฟล์ไม่มีแล้วออก และ `docker compose up -d postgres redis` ทำให้ทั้งสอง service เป็น healthy (คนรันตรวจ)
- สถานะ: เสร็จ รอทีมตรวจ (coordinator ตรวจ diff และสถานะ container ซ้ำแล้ว 8 พฤศจิกายน 2569)

### T-03 สร้าง migration ตาราง members และตัวช่วย test ฐานข้อมูล
- รองรับ: CON-DB-01, FR-AUTH-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-04, T-06
- ไฟล์ที่แตะ: `backend/migrations/0001_create_members.up.sql`, `backend/migrations/0001_create_members.down.sql`, `backend/internal/testutil/fixtures.go`, `backend/internal/testutil/fixtures_test.go`, `backend/go.mod`, `backend/go.sum`
- ต้องทำหลัง: T-01, T-02
- เสร็จเมื่อ: migrate up สร้างตาราง `members` (`id`, `student_id` UNIQUE, `password_hash`, `created_at`) และ migrate down ลบได้ ยืนยันด้วย `TestMembersMigration` บน PostgreSQL จริง; helper `CreateMember(...)` สร้างสมาชิกจำลองด้วย bcrypt hash สำหรับ test เท่านั้น (ไม่มีโค้ด production สร้างสมาชิก ตาม Q-03)
- สถานะ: เสร็จ รอทีมตรวจ (coordinator รัน test ซ้ำและรันทั้งชุดพร้อมกัน 3 รอบผ่านหมด ฐานข้อมูลสะอาดแล้ว 8 พฤศจิกายน 2569; bcrypt ยังรอทีมยืนยันตาม plan D5)

### T-04 เขียน member store (ค้นสมาชิกตามรหัสนักศึกษา)
- รองรับ: FR-AUTH-01, CON-DB-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-06, T-07
- ไฟล์ที่แตะ: `backend/internal/auth/member_store.go`, `backend/internal/auth/member_store_test.go`
- ต้องทำหลัง: T-03
- เสร็จเมื่อ: `FindByStudentID` คืนสมาชิก (id, student_id, password_hash) เมื่อมี และคืนผลว่าไม่พบเมื่อไม่มี ยืนยันด้วย test กับ PostgreSQL จริง
- สถานะ: เสร็จ รอทีมตรวจ (coordinator รัน test ซ้ำทั้งชุดแล้ว 8 พฤศจิกายน 2569; หมายเหตุ: package doc comment ซ้ำกับ session_store.go เป็นเรื่องรูปแบบ)

### T-05 เขียน session store (Redis): สร้าง token สุ่ม, hash, เก็บพร้อม TTL 7 วัน, อ่าน
- รองรับ: CON-CACHE-01, FR-AUTH-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-06, T-08, T-09
- ไฟล์ที่แตะ: `backend/internal/auth/session_store.go`, `backend/internal/auth/session_store_test.go`, `backend/go.mod`, `backend/go.sum`
- ต้องทำหลัง: T-01, T-02
- เสร็จเมื่อ: `Create` สร้าง token จาก `crypto/rand` เก็บ key `session:<sha256-hex>` ค่า `{member_id, student_id}` และ TTL ปกติ = 7 วัน (TTL กำหนดจากภายนอกได้เพื่อทดสอบ) `Get` คืนข้อมูลหรือผลว่าไม่พบ; การสร้างสองครั้งได้ key คนละอัน; test กับ Redis จริงผ่าน
- สถานะ: เสร็จ รอทีมตรวจ (coordinator รัน test ซ้ำแล้ว 8 พฤศจิกายน 2569; go.mod เพิ่ม go-redis v9.23.0 และ cpuid ถูกอัปเกรดเป็น indirect)

### T-06 ทำ login: ตรวจรหัสผ่านด้วย bcrypt แล้วออก Refresh Token
- รองรับ: FR-AUTH-01, CON-CACHE-01
- ตรวจด้วย: AC-AUTH-01, AC-AUTH-05
- ไฟล์ที่แตะ: `backend/internal/auth/service.go`, `backend/internal/auth/handler.go`, `backend/internal/app/router.go`, `backend/cmd/api/main.go`, `backend/internal/auth/login_test.go`, `backend/go.mod`, `backend/go.sum` (แก้รอบที่ 2: เพิ่ม `cmd/api/main.go` เพื่อต่อ PostgreSQL/Redis เข้ากับ handler จริง ไม่เช่นนั้น T-16 รัน API จริงไม่ได้)
- ต้องทำหลัง: T-01, T-03, T-04, T-05
- เสร็จเมื่อ: `TestAC_AUTH_01` และ `TestAC_AUTH_05` ผ่าน (`POST /api/auth/login` ด้วยรหัสถูกต้องได้ 200 พร้อม `refresh_token` และ `expires_in`; คำนวณ SHA-256 ของ token แล้วพบ key `session:<hash>` ใน Redis จริงพร้อม `member_id` ตรงกัน)
- สถานะ: เสร็จ รอทีมตรวจ (coordinator อ่านโค้ดและรัน test ทั้งชุดซ้ำแล้ว 8 พฤศจิกายน 2569; ไม่ได้ curl login สำเร็จกับเซิร์ฟเวอร์จริงเพราะต้องแทรกสมาชิกลง DB ร่วม; ข้าม T-11 จึง error ของ dependency ทุกชนิดตอบ 503 และไม่มี timeout ที่เรากำหนดเอง)

### T-07 ทำให้การปฏิเสธ login ไม่บอกว่าผิดข้อใด
- รองรับ: FR-AUTH-01
- ตรวจด้วย: AC-AUTH-06
- ไฟล์ที่แตะ: `backend/internal/auth/service.go`, `backend/internal/auth/handler.go`, `backend/internal/auth/login_failure_test.go`
- ต้องทำหลัง: T-06
- เสร็จเมื่อ: `TestAC_AUTH_06` ผ่าน (รหัสนักศึกษาที่ไม่มีในระบบ และรหัสผ่านผิด ได้ 401 ที่ body เหมือนกันทุกตัวอักษร ไม่มี `refresh_token`; กรณีไม่พบสมาชิกยังเทียบ bcrypt กับ hash หลอกเพื่อให้เวลาตอบใกล้เคียงกัน)
- สถานะ: เสร็จ รอทีมตรวจ (coordinator อ่านโค้ดและรัน test AC_AUTH_06 ซ้ำแล้ว 8 พฤศจิกายน 2569; handler.go ไม่ต้องแก้จึงไม่ถูกแตะ; cost ของ hash หลอก = bcrypt.DefaultCost เป็นการเลือกของทีม และ test ด้านเวลาเป็นแค่ตัวกัน หลักฐานจริงคือ test แบบนับการเรียก bcrypt)

### T-08 ทำ GET /api/auth/session และ middleware RequireSession
- รองรับ: FR-AUTH-01, CON-CACHE-01
- ตรวจด้วย: AC-AUTH-02, AC-AUTH-04
- ไฟล์ที่แตะ: `backend/internal/auth/middleware.go`, `backend/internal/auth/handler.go`, `backend/internal/app/router.go`, `backend/cmd/api/main.go`, `backend/internal/auth/session_test.go` (แก้รอบที่ 3: เพิ่ม `cmd/api/main.go` เพื่อต่อตัวอ่าน session เข้ากับ handler จริง)
- ต้องทำหลัง: T-06
- เสร็จเมื่อ: `TestAC_AUTH_02` และ `TestAC_AUTH_04` ผ่าน (token ที่ออกให้ใช้กับ `GET /api/auth/session` ได้ 200 ได้ `member.id` ถูกต้อง, route ทดสอบที่ห่อด้วย `RequireSession` เรียกได้โดยไม่ส่งรหัสผ่าน, TTL ของ key ใน Redis อยู่ในช่วง 7 วัน, token สุ่มและ token ที่แก้ 1 ตัวอักษรได้ 401) ตาม plan D2 endpoint นี้ไม่เรียก PostgreSQL
- สถานะ: เสร็จ รอทีมตรวจ (coordinator อ่านโค้ดและรัน test ซ้ำแล้ว 8 พฤศจิกายน 2569; D2 ยังรอทีมตัดสิน: PostgreSQL ล่มแล้ว endpoint นี้ยังตอบ 200)

### T-09 ทดสอบ token หมดอายุ
- รองรับ: FR-AUTH-01
- ตรวจด้วย: AC-AUTH-03
- ไฟล์ที่แตะ: `backend/internal/auth/session_expiry_test.go`
- ต้องทำหลัง: T-08
- เสร็จเมื่อ: `TestAC_AUTH_03` ผ่าน (สร้าง service ด้วย TTL สั้นใน test ล็อกอิน รอเกิน TTL แล้ว `GET /api/auth/session` ได้ 401)
- สถานะ: เสร็จ รอทีมตรวจ (coordinator รัน test ซ้ำและทำ mutation check แล้ว 8 พฤศจิกายน 2569: ทำให้ production ไม่ตั้งอายุ → TestAC_AUTH_03 ล้มจริง; test นี้พิสูจน์ "กลไกหมดอายุ" ด้วย TTL 1 วินาที ส่วนค่า 7 วันพิสูจน์ที่ TestAC_AUTH_02)

### T-10 ทดสอบการเข้าสู่ระบบหลายอุปกรณ์
- รองรับ: FR-AUTH-01
- ตรวจด้วย: AC-AUTH-09
- ไฟล์ที่แตะ: `backend/internal/auth/multi_device_test.go`
- ต้องทำหลัง: T-08
- เสร็จเมื่อ: `TestAC_AUTH_09` ผ่าน (เข้าสู่ระบบสองครั้งได้ token ต่างกัน ทั้งสอง token เรียก `GET /api/auth/session` ได้ 200)
- สถานะ: เสร็จ รอทีมตรวจ (coordinator รัน test ซ้ำและทำ mutation check แล้ว 8 พฤศจิกายน 2569: ทำให้ token ซ้ำกันทุกเครื่อง → TestAC_AUTH_09 ล้มจริง)

### T-11 ทำกลไกตอบ 503 เมื่อ Redis หรือ PostgreSQL ไม่ตอบ และรับ timeout จาก config
- รองรับ: CON-CACHE-01, CON-DB-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-12
- ไฟล์ที่แตะ: `backend/internal/config/config.go`, `backend/internal/auth/session_store.go`, `backend/internal/auth/member_store.go`, `backend/internal/auth/handler.go`, `backend/internal/auth/middleware.go`, `backend/internal/auth/unavailable_test.go`
- ต้องทำหลัง: T-08
- เสร็จเมื่อ: config อ่านค่า timeout จาก env โดย **ไม่มีค่าเริ่มต้น** (ถ้าไม่ตั้งค่า โปรแกรมไม่เริ่มและบอกสาเหตุ) และ handler คืน 503 `service_unavailable` พร้อมข้อความให้ลองใหม่ เมื่อ store ปลอมคืน error ประเภท timeout/เชื่อมต่อล้มได้ ยืนยันด้วย unit test
- สถานะ: ข้าม (ทีมสั่ง 8 พฤศจิกายน 2569 ให้ข้ามงาน blackhole) — เดิมรอ Q-04
- **อัปเดต 8 พฤศจิกายน 2569 (หลัง API test พบ F-1):** ทำแล้วโดย sub-agent แต่แนวทางต่างจากข้อกำหนดเดิมของ task นี้: ใช้ context deadline ต่อคำขอที่ `handler.go`/`middleware.go` (ไม่ได้ทำที่ store) และ `config` มีค่าเริ่มต้น 5 วินาที (ASM-09 ที่ทีมเลือกเมื่อ 8 พฤศจิกายน 2569) แทน "ไม่มีค่าเริ่มต้น" ไฟล์ที่แก้: config.go, handler.go, middleware.go, cmd/api/main.go + test ใหม่ config_test.go, dependency_timeout_test.go — **สถานะ: เสร็จ รอทีมตรวจ** (coordinator อ่านโค้ดแล้ว; ยังไม่ได้ยิงซ้ำด้วย curl/docker pause กับระบบจริงตามที่ทีมสั่งให้ข้ามการทดสอบ)

### T-12 ทดสอบเมื่อ Redis หรือ PostgreSQL ไม่ตอบสนอง (ตัวจำลอง blackhole)
- รองรับ: CON-CACHE-01, CON-DB-01
- ตรวจด้วย: AC-AUTH-07, AC-AUTH-08
- ไฟล์ที่แตะ: `backend/internal/testutil/blackhole.go`, `backend/internal/auth/dependency_failure_test.go`
- ต้องทำหลัง: T-11
- เสร็จเมื่อ: `TestAC_AUTH_07` (Redis blackhole, ขอใช้งานต่อด้วย token ถูกต้อง → 503) และ `TestAC_AUTH_08` (PostgreSQL blackhole, เข้าสู่ระบบ → 503) ผ่าน ภายในเวลา timeout ที่ test กำหนดเอง
- สถานะ: ข้าม (ทีมสั่ง 8 พฤศจิกายน 2569 ให้ข้ามงาน blackhole) — เดิมรอ Q-04; **AC-AUTH-07 และ AC-AUTH-08 จึงยังไม่มี test**
- **อัปเดต 8 พฤศจิกายน 2569:** ยังข้าม blackhole (เครือข่ายจริง) ตามคำสั่งทีม แต่ AC-AUTH-07/08 มี test แบบ blocking fake ที่บล็อกจน context หมดเวลาแล้ว (`TestAC_AUTH_07`, `TestAC_AUTH_08` ใน `dependency_timeout_test.go` สร้างโดย fixer ของ F-1) ซึ่งพิสูจน์ได้แค่ว่าชั้น HTTP ตั้ง deadline และตอบ 503 ไม่ใช่พฤติกรรมบนเครือข่ายจริง หลักฐานบนเครือข่ายจริงมีเฉพาะตอนก่อนแก้ (API_test.md I-08: ค้าง) ยังไม่มีหลักฐานหลังแก้

---

## ฝั่ง Frontend (React Vite) — ใช้ API จำลองตามสัญญา plan §4.1 ไม่ต้องรอ task ฝั่ง backend

### T-13 สร้างโครง frontend และ API จำลอง (msw)
- รองรับ: CON-TECH-02
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-14, T-15
- ไฟล์ที่แตะ: `frontend/package.json`, `frontend/package-lock.json`, `frontend/vite.config.ts`, `frontend/index.html`, `frontend/tsconfig.json`, `frontend/src/vite-env.d.ts`, `frontend/src/main.tsx`, `frontend/src/test/setup.ts`, `frontend/src/mocks/handlers.ts`, `frontend/src/mocks/server.ts`, `frontend/src/mocks/handlers.test.ts` (แก้รอบที่ 1: เพิ่ม 5 ไฟล์ที่โครง Vite/vitest ต้องมี)
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: `npm test` รัน test ที่ยืนยันว่า msw ตอบ `POST /api/auth/login` และ `GET /api/auth/session` ตามสัญญา (200/401/503) ได้ผ่าน
- สถานะ: เสร็จ รอทีมตรวจ (coordinator รัน npm test 9/9 และ npm run build ซ้ำแล้ว 8 พฤศจิกายน 2569; ยังไม่ได้ยืนยันด้วยการรันว่า onUnhandledFrame:"error" ทำให้ request ที่ไม่มี handler ล้มจริง)

### T-14 ทำ API client, AuthProvider และ RequireAuth
- รองรับ: FR-AUTH-01, CON-TECH-02
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-15
- ไฟล์ที่แตะ: `frontend/src/features/auth/api.ts`, `frontend/src/features/auth/AuthProvider.tsx`, `frontend/src/features/auth/RequireAuth.tsx`, `frontend/src/features/auth/AuthProvider.test.tsx`
- ต้องทำหลัง: T-13
- เสร็จเมื่อ: test ผ่านว่าเมื่อเริ่มแอปด้วย token ใน `localStorage`: ตอบ 200 → สถานะเข้าสู่ระบบ, ตอบ 401 → ลบ token และไป `/login`, ตอบ 503/เครือข่ายล้ม → แสดงให้ลองใหม่ และ **ไม่ลบ token**
- สถานะ: เสร็จ รอทีมตรวจ (coordinator รัน npm test 18/18 และ tsc ซ้ำแล้ว 8 พฤศจิกายน 2569; fetch ฝั่งเว็บยังไม่มี timeout ซึ่งผูกกับ Q-04)

### T-15 ทำ LoginPage, HomePage และทดสอบการใช้งานต่อหลังเปิดเว็บใหม่
- รองรับ: FR-AUTH-01, CON-TECH-02
- ตรวจด้วย: AC-AUTH-10
- ไฟล์ที่แตะ: `frontend/src/features/auth/LoginPage.tsx`, `frontend/src/features/auth/HomePage.tsx`, `frontend/src/features/auth/AC_AUTH_10.test.tsx`, `frontend/src/App.tsx`, `frontend/src/main.tsx` (แก้รอบที่ 1: เพิ่ม App.tsx และ main.tsx เพื่อต่อ AuthProvider/router เข้ากับแอป)
- ต้องทำหลัง: T-14
- เสร็จเมื่อ: test `AC_AUTH_10: resume after reload` ผ่าน (ใส่ token ไว้ใน `localStorage` แล้ว mount แอปใหม่ โดย msw ตอบ 200 → เห็นรหัสนักศึกษาบน `HomePage` ไม่เห็นฟอร์มเข้าสู่ระบบ; ไม่มีปุ่มออกจากระบบ)
- สถานะ: เสร็จ รอทีมตรวจ (coordinator อ่านโค้ดและรัน npm test 25/25 กับ tsc ซ้ำแล้ว 8 พฤศจิกายน 2569; test จำลองการปิดเบราว์เซอร์ด้วย unmount ส่วนเบราว์เซอร์จริงอยู่ที่ T-16)

### T-16 ต่อ frontend กับ API จริงและตรวจด้วยมือ
- รองรับ: FR-AUTH-01, CON-TECH-02
- ตรวจด้วย: AC-AUTH-10 (ส่วนที่ต้องปิดเบราว์เซอร์จริง)
- ไฟล์ที่แตะ: `frontend/vite.config.ts`, `frontend/src/features/auth/api.ts`
- ต้องทำหลัง: T-08, T-15
- เสร็จเมื่อ: ผู้ตรวจเข้าสู่ระบบจากเว็บกับ backend จริงด้วยสมาชิกทดสอบ (ใส่ด้วยมือผ่าน psql ใช้ bcrypt hash ที่สร้างจากเครื่องมือภายนอก ไม่เพิ่มโค้ด seed ตาม Q-03) ปิดเบราว์เซอร์ทั้งโปรแกรมแล้วเปิดเว็บใหม่ภายใน 7 วัน ยังเห็นหน้าหลักโดยไม่ต้องเข้าสู่ระบบ และบันทึกผลใน `prompt-log.md`
- สถานะ: พร้อมทำ

---

## ตารางตรวจความครบ

### AC → task ที่ตรวจ
| AC ID | task ที่ตรวจ AC นี้ |
|---|---|
| AC-AUTH-01 | T-06 |
| AC-AUTH-02 | T-08 |
| AC-AUTH-03 | T-09 |
| AC-AUTH-04 | T-08 |
| AC-AUTH-05 | T-06 |
| AC-AUTH-06 | T-07 |
| AC-AUTH-07 | T-12 (รอ Q-04) |
| AC-AUTH-08 | T-12 (รอ Q-04) |
| AC-AUTH-09 | T-10 |
| AC-AUTH-10 | T-15, T-16 |

### Constraint → task ที่ทำให้เป็นจริง
| Constraint ID | task ที่ทำให้เป็นจริง |
|---|---|
| CON-CACHE-01 | T-02 (Redis ขึ้นได้), T-05 (session store), T-06 (AC-AUTH-05), T-08, T-11, T-12 |
| CON-DB-01 | T-02 (PostgreSQL ขึ้นได้), T-03 (ตาราง members), T-04 (member store), T-11, T-12 |
| CON-TECH-01 | T-01 และทุก task ฝั่ง backend (T-03 ถึง T-12) |
| CON-TECH-02 | T-13, T-14, T-15, T-16 |
| DOM-*, IF-* | ไม่มีใน spec นี้ |

## สิ่งที่ยังไม่ทำ (Open Questions จาก spec)
| รหัส | คำถาม | task ที่รอ |
|---|---|---|
| Q2 (SRS) | กลุ่มสมาชิกที่โดน Auto-logout และ Session ต้องคงอยู่นานเท่าใด — ทีมยืนยัน ASM-02 แล้ว (ยังไม่ได้ถามเจ้าของเรื่องจริง) | ไม่มี task รอ (T-05, T-08, T-09 ใช้ 7 วันตาม ASM-02 ถ้าตอบต่างต้องแก้ task เหล่านี้) |
| Q7 (SRS) | ผู้ใช้บริการมีเฉพาะนักศึกษาหรือไม่ — ทีมยืนยัน ASM-01 แล้ว (ยังไม่ได้ถามเจ้าของเรื่องจริง) | ไม่มี task รอ (T-03, T-06 ใช้รหัสนักศึกษา+รหัสผ่านตาม ASM-01) |
| Q-01 | อัตราการหลุดโดยไม่ตั้งใจสูงสุดที่ยอมรับได้ (NFR-REL-01) เท่าใด — ทีมตอบ: ไม่รู้ ต้องถาม Product Owner | ยังไม่มี task (NFR-REL-01 ไม่มี AC จึงไม่มี task ทดสอบ จะเพิ่มเมื่อได้คำตอบ) |
| Q-02 | ปิดแล้วสำหรับ sprint นี้: ไม่มี logout (ASM-03) | ไม่มี task รอ (ไม่มี task ออกจากระบบ) |
| Q-03 | บัญชีสมาชิกและรหัสผ่านเริ่มต้นมาจากไหน — ทีมตอบ: ไม่รู้ ต้องถาม | ยังไม่มี task สร้าง/นำเข้าสมาชิกในโค้ด production (ใช้ข้อมูลจำลองใน test และตอนตรวจ T-16 ด้วยมือเท่านั้น) |
| Q-04 | กี่วินาทีที่ถือว่า Redis หรือ PostgreSQL "ไม่ตอบสนอง" | **T-11, T-12 รออยู่** |
