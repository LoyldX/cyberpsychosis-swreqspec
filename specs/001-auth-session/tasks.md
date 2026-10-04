# Tasks: รักษา Session Login แบบ Persistent
- Feature: รักษา Session Login แบบ Persistent (`001-auth-session`)
- Spec ID: FR-AUTH-01, NFR-REL-01, CON-CACHE-01 (AC ทั้งหมดเป็น Candidate)
- อ้างอิง plan.md: `specs/001-auth-session/plan.md`
- วันที่: 4 ตุลาคม 2569
- สรุป: ทั้งหมด 6 task, รอ Q-xx 4 task (พร้อมทำ 2 task). ยังไม่มี task ใดถูกเริ่มทำ
- หมายเหตุ: ไฟล์ตามโครงค่าเริ่มต้น (React Vite / FastAPI) ทีมเลือกเอง ไม่ได้มาจาก spec

## รายการ task

### T-01 สร้างโครงข้อมูล session ใน Redis
- รองรับ: FR-AUTH-01, CON-CACHE-01
- ตรวจด้วย: AC-AUTH-03 (Candidate)
- ไฟล์ที่แตะ: `backend/app/models/session.py`, `backend/app/repositories/session_store.py`, `backend/tests/test_auth_session.py`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: `test_AC_AUTH_03_session_stored_in_redis` ผ่าน (มี key ใน Redis หลังสร้าง session) โดยยังไม่ตั้งค่า TTL
- สถานะ: พร้อมทำ

### T-02 สร้าง endpoint ต่ออายุ session ด้วย refresh token
- รองรับ: FR-AUTH-01, CON-CACHE-01
- ตรวจด้วย: AC-AUTH-01 (Candidate)
- ไฟล์ที่แตะ: `backend/app/api/auth.py`, `backend/tests/test_auth_session.py`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: `test_AC_AUTH_01_persistent_session_reuses_refresh_token` ผ่าน ค่าอายุ session ตาม Q2 เท่านั้น
- สถานะ: รอ Q2 (อายุ session และกลุ่มผู้ใช้ "User" ยังไม่มีคำตอบ)

### T-03 สร้าง endpoint ออกจากระบบ
- รองรับ: FR-AUTH-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ (งานพื้นฐาน) ตรวจว่า session ถูกลบออกจาก Redis
- ไฟล์ที่แตะ: `backend/app/api/auth.py`, `backend/tests/test_auth_session.py`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: test ยืนยันว่า DELETE `/auth/session` ลบ key ใน Redis แล้วการเรียก refresh ซ้ำไม่สำเร็จ
- สถานะ: พร้อมทำ

### T-04 วัดอัตราการหลุดของ session
- รองรับ: NFR-REL-01, CON-CACHE-01
- ตรวจด้วย: AC-AUTH-02 (Candidate)
- ไฟล์ที่แตะ: `backend/tests/test_auth_reliability.py`
- ต้องทำหลัง: T-02
- เสร็จเมื่อ: `test_AC_AUTH_02_unintended_logout_rate_within_threshold` ผ่านเกณฑ์ที่ทีมกำหนด (เกณฑ์ = TBD)
- สถานะ: รอ Q2 (อัตราหลุดสูงสุดยังไม่มีค่า)

### T-05 สร้างหน้าจอเข้าใช้งานต่อเนื่อง
- รองรับ: FR-AUTH-01 (ข้อเสนอ UI-AUTH-01 เป็น Candidate)
- ตรวจด้วย: ไม่มี AC ที่ยอมรับแล้ว (UI-AUTH-01 เป็น Candidate)
- ไฟล์ที่แตะ: `frontend/src/pages/AutoLogin.jsx`, `frontend/src/api/mockAuth.js`
- ต้องทำหลัง: ไม่มี (ใช้ API จำลองตามสัญญาใน plan.md ข้อ 4)
- เสร็จเมื่อ: หน้าจอเปิดเข้าหน้าหลักทันทีเมื่อ API จำลองตอบว่า session ยังใช้ได้
- สถานะ: รอ Q-xx (UI-AUTH-01 ยังไม่มีหมายเลข Q; ที่เกี่ยวข้องคือ Q2)

### T-06 ต่อหน้าจอกับ API จริง
- รองรับ: FR-AUTH-01, CON-CACHE-01
- ตรวจด้วย: AC-AUTH-01 (Candidate)
- ไฟล์ที่แตะ: `frontend/src/api/auth.js`, `frontend/src/pages/AutoLogin.jsx`
- ต้องทำหลัง: T-02, T-05
- เสร็จเมื่อ: เปิดแอปครั้งที่สองภายในอายุ session แล้วเข้าหน้าหลักได้โดยไม่ล็อกอินซ้ำ
- สถานะ: รอ Q2

## ตารางตรวจความครบ

### AC ทั้งหมด (Candidate)
| AC ID | task ที่ตรวจ AC นี้ |
|---|---|
| AC-AUTH-01 (Candidate) | T-02, T-06 |
| AC-AUTH-02 (Candidate) | T-04 |
| AC-AUTH-03 (Candidate) | T-01 |

### Constraint ทั้งหมด
| Constraint ID | task ที่ทำให้เป็นจริง |
|---|---|
| CON-CACHE-01 | T-01, T-02, T-04 |

## สิ่งที่ยังไม่ทำ
- **Q2** "User" ที่โดน Auto-logout คือผู้ใช้กลุ่มใด และ session ต้องคงอยู่นานเท่าใด: รอ T-02, T-04, T-06 (ค่าอายุ session, เกณฑ์อัตราหลุด และการต่อหน้าจอ)
- UI-AUTH-01 (Candidate): รอ T-05 และ T-06 (ยังไม่มีหมายเลข Q โดยตรง)
