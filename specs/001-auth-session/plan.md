# แผน: รักษา Session Login แบบ Persistent
Spec: `specs/001-auth-session/spec.md` (Draft v2 — ยังไม่ผ่าน clarify) | สร้างเมื่อ 4 ตุลาคม 2569 | หน้านี้เป็นผลลัพธ์ที่ AI ร่าง คนตรวจต้องยืนยัน

## 1. สรุปแนวทาง
ระบบรักษา session ของสมาชิกด้วย refresh token ตาม FR-AUTH-01 และเก็บข้อมูล session ใน Redis ตาม CON-CACHE-01
เป้าหมายคือให้สมาชิกไม่ต้องล็อกอินซ้ำโดยไม่จำเป็น (ลดการ Auto-logout) ตาม pain point ใน refine-req §1
เกณฑ์ความน่าเชื่อถือ (อัตราการหลุด) ตาม NFR-REL-01 ยังวัดไม่ได้จนกว่าจะตอบ Q2
อายุ session และกลุ่มผู้ใช้ "User" ยังไม่กำหนด จึงยังไม่ใส่ค่า TTL หรือเงื่อนไขกลุ่มผู้ใช้ในโค้ด
ส่วนหน้าจอ (UI-AUTH-01 เป็น Candidate) ยังไม่ถูกสร้างจริงจนกว่าทีมยืนยัน

## 2. เทคโนโลยีที่ใช้
| สิ่งที่เลือก | มาจาก | หมายเหตุ |
|---|---|---|
| Redis | CON-CACHE-01 | เก็บข้อมูล session ตามข้อกำหนด |
| Refresh token | FR-AUTH-01, NFR-REL-01 | รูปแบบ token ยังไม่ออกแบบ (ไม่มีใน spec) |
| หน้าบ้าน: React (Vite) | ทีมเลือกเอง ไม่ได้มาจาก spec | ใช้เฉพาะหน้าจอ UI-AUTH-01 (Candidate) |
| หลังบ้าน: Python FastAPI | ทีมเลือกเอง ไม่ได้มาจาก spec | ค่าเริ่มต้นของรายวิชา |

## 3. โมเดลข้อมูล
| Entity | ฟิลด์หลัก | รองรับ |
|---|---|---|
| Session (ใน Redis) | session identifier, member reference, refresh token state, expiration state | FR-AUTH-01, NFR-REL-01, CON-CACHE-01 |

- ค่า expiration ยังไม่กำหนด รอ Q2 (โครงฟิลด์มีแต่ค่าไม่ถูกตั้ง)
- ไม่มีตาราง relational สำหรับ session เพราะ spec ระบุ Redis

## 4. API / หน้าจอ
| รายการ | Input / Output หลัก | รองรับ |
|---|---|---|
| POST `/auth/refresh` | refresh token / session ใหม่ | FR-AUTH-01 |
| DELETE `/auth/session` | session identifier / ผลการออกจากระบบ | FR-AUTH-01 |
| หน้าจอเข้าใช้งานต่อเนื่อง (UI-AUTH-01, Candidate) | สถานะ session / เข้าหน้าหลักทันที | FR-AUTH-01, NFR-REL-01 |

## 5. ตารางตรวจ Constraints
| Constraint ID | ถูกนำไปใช้ที่ไหนใน plan | สถานะ |
|---|---|---|
| CON-CACHE-01 | โมเดล Session (ข้อ 3) และ POST `/auth/refresh` (ข้อ 4) ใช้ Redis เป็นที่เก็บ | ใช้แล้ว |
| CON-DB-01 | ไม่เกี่ยวข้องกับ spec นี้ | ยังไม่ได้ใช้ เพราะ refine-req ไม่ได้ระบุ session ใน relational DB |

## 6. แผนทดสอบจาก Acceptance Criteria
ทุก AC เป็น Candidate (ยังไม่ผ่านการยืนยันจากทีม)

| AC ID | ชื่อ test | ทดสอบอย่างไร |
|---|---|---|
| AC-AUTH-01 (Candidate) | `test_AC_AUTH_01_persistent_session_reuses_refresh_token` | สร้าง session, เรียก refresh ภายในอายุ session แล้วตรวจว่าเข้าใช้งานได้โดยไม่ล็อกอินซ้ำ (ค่าอายุรอ Q2) |
| AC-AUTH-02 (Candidate) | `test_AC_AUTH_02_unintended_logout_rate_within_threshold` | จำลองการใช้งานต่อเนื่องและนับ session หลุด เทียบกับเกณฑ์ (เกณฑ์ = TBD รอ Q2) |
| AC-AUTH-03 (Candidate) | `test_AC_AUTH_03_session_stored_in_redis` | สร้าง session แล้วตรวจว่ามี key ใน Redis ไม่ใช่ที่เก็บอื่น |

## 7. ลำดับงาน
1. วาง Session ใน Redis โดยยังไม่ตั้ง TTL (CON-CACHE-01, AC-AUTH-03)
2. สร้าง endpoint refresh token (FR-AUTH-01, AC-AUTH-01 รอ Q2 ด้านอายุ)
3. สร้าง endpoint ออกจากระบบ (FR-AUTH-01)
4. ออกแบบวิธีวัดอัตราการหลุด (NFR-REL-01, AC-AUTH-02 รอ Q2)
5. สร้างหน้าจอเข้าใช้งานต่อเนื่อง (UI-AUTH-01 Candidate, รอทีมยืนยัน)
6. เชื่อมหน้าจอกับ API จริงและรัน test ทั้งหมด (รอ Q2 สำหรับ AC-AUTH-01/02)

## 8. สิ่งที่ยังไม่ทำ
- **Q2** "User" ที่โดน Auto-logout คือผู้ใช้กลุ่มใด และ session ต้องคงอยู่นานเท่าใด
  ส่วนที่เกี่ยวข้องกับข้อนี้จะยังไม่สร้างจนกว่าจะได้คำตอบ (อายุ session, เกณฑ์อัตราหลุด, UI-AUTH-01)
