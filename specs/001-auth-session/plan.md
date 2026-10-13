# แผน: Auth Session Persistence
Spec: `specs/001-auth-session/spec.md` (Draft v1) | สร้างเมื่อ 13 ตุลาคม 2569 | หน้านี้เป็นผลลัพธ์ที่ AI ร่าง คนตรวจต้องยืนยัน

## 1. สรุปแนวทาง
ระบบต้องรองรับการเข้าสู่ระบบของสมาชิกด้วยรหัสนักศึกษา/SSO เป็นหลัก และต้องยืนยัน Session ให้คงอยู่ผ่าน Refresh Token โดยใช้ Redis จัดการ token state และยืนยันว่า Refresh Token มีอายุ 7 วัน
การกำหนดค่า TTL 7 วัน ถือเป็นความชัดเจนที่ยอมรับแล้วจากการปรับ spec ล่าสุด
เงื่อนไขเรื่องอัตราการหลุดโดยไม่ตั้งใจและค่ากำหนดดัชนีความน่าเชื่อถือยังไม่ได้ระบุเป็นตัวเลขชัดเจน จึงยังไม่มี AC ให้ทดสอบแบบเชิงตัวเลข

## 2. เทคโนโลยีที่ใช้
| สิ่งที่เลือก | มาจาก | หมายเหตุ |
|---|---|---|
| Redis | CON-CACHE-01 | บริหาร Session Login และ Refresh Token state |
| ระบบยืนยันตัวตนสมาชิก | ทีมเลือกเอง / ขยายจาก FR-AUTH-01 | การเข้าสู่ระบบหลักใช้รหัสนักศึกษา/SSO |
| หลังบ้าน: โครงโปรเจกต์ที่มีอยู่ใน repo | ทีมเลือกเอง ไม่ได้มาจาก spec | กำหนดตาม stack ของโปรเจกต์ปัจจุบัน |

## 3. โมเดลข้อมูล
| Entity | ฟิลด์หลัก | รองรับ |
|---|---|---|
| User session | `user_id`, `session_id`, `refresh_token`, `expires_at`, `issued_at`, `status` | FR-AUTH-01, NFR-REL-01 |
| Refresh token record | `token_hash`, `user_id`, `expires_at`, `revoked_at` | FR-AUTH-01, CON-CACHE-01 |

- ฟิลด์เพิ่มเติม เช่น device metadata, IP, last_seen จะเพิ่มได้ก็ต่อเมื่อทีมยืนยันว่าพวกนั้นจำเป็น
- TTL ควรตั้งที่ 7 วัน สำหรับ Refresh Token ตามข้อยืนยันของ feature นี้

## 4. API / หน้าจอ
| รายการ | Input / Output หลัก | รองรับ |
|---|---|---|
| POST `/auth/login` | รหัสนักศึกษา/SSO หรือข้อมูลประจำตัวผู้ใช้ / access token + refresh token | FR-AUTH-01 |
| POST `/auth/refresh` | refresh token / access token ใหม่ หรือ reject หาก token หมดอายุ/ถูกยกเลิก | FR-AUTH-01, NFR-REL-01 |
| POST `/auth/logout` | refresh token / ยกเลิก session | FR-AUTH-01 |
| GET `/auth/session` | สถานะ Session ปัจจุบัน / logged-in หรือ logged-out | FR-AUTH-01 |

## 5. ตารางตรวจ Constraints
| Constraint ID | ถูกนำไปใช้ที่ไหนใน plan | สถานะ |
|---|---|---|
| CON-CACHE-01 | Redis สำหรับจัดการ session และ refresh token | ใช้จริง |

## 6. แผนทดสอบ
ยังไม่มี AC ใน spec ที่ระบุเกณฑ์เชิงตัวเลขสำหรับอัตราการหลุดโดยไม่ตั้งใจ จึงยังไม่สามารถเขียนเกณฑ์เชิงตัวเลขได้ชัดเจน
สมมติฐาน: จนกว่าทีมจะกำหนด AC ให้ตั้งชื่อ test ตาม ID ที่มีใน spec เช่น `test_FR_AUTH_01_...` และไม่อ้าง AC ใด

## 7. ลำดับงาน
1. T-01 สร้างโครงข้อมูล session และ refresh token (รองรับ: FR-AUTH-01, NFR-REL-01, CON-CACHE-01)
2. T-02 สร้าง endpoint ล็อกอินด้วยรหัสนักศึกษา/SSO และสร้าง refresh token พร้อม TTL 7 วัน (รองรับ: FR-AUTH-01)
3. T-03 สร้าง endpoint refresh token และตรวจสอบ expiry / revoke (รองรับ: FR-AUTH-01, NFR-REL-01)
4. T-04 เชื่อม Redis สำหรับจัดเก็บ session state และยืนยันการบันทึกค่า TTL (รองรับ: CON-CACHE-01)
5. T-05 สร้าง endpoint logout และยกเลิก token ที่ใช้อยู่ (รองรับ: FR-AUTH-01)

## 8. สิ่งที่ยังไม่ทำ
- ยังไม่มีหมายเลข Q: เกณฑ์อัตราการหลุดโดยไม่ตั้งใจของ NFR-REL-01 ยังไม่กำหนดตัวเลข
- ยังไม่มีหมายเลข Q: สามารถใช้ SSO หลาย provider หรือเพียงแค่รหัสนักศึกษา/SSO ตัวเดียวเท่านั้น
- ยังไม่มีหมายเลข Q: จะมีการ refresh token แบบ rotation หรือ reuse ตามกลยุทธ์เดียวกัน
- ยังไม่มีหมายเลข Q: กลุ่มสมาชิกที่มีสถานะพิเศษหรือไม่ใช่นักศึกษาจะเข้าระบบอย่างไร
