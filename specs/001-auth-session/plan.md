# แผน: รักษา Session Login แบบ Persistent

## 1. สรุปแนวทาง

ระบบจะรักษา session ของสมาชิกด้วย refresh token ตาม FR-AUTH-01
ข้อมูล session จะจัดการผ่าน Redis ตาม CON-CACHE-01
ระบบจะรองรับการต่ออายุ session โดยไม่บังคับสมาชิกเข้าสู่ระบบใหม่โดยไม่จำเป็น
เกณฑ์ความน่าเชื่อถือจะตรวจตาม NFR-REL-01
อายุ session และกลุ่มผู้ใช้ยังไม่กำหนดจนกว่าจะตอบ Q2

## 2. เทคโนโลยีที่ใช้

| สิ่งที่เลือก | มาจาก | หมายเหตุ |
|---|---|---|
| Redis | CON-CACHE-01 | ใช้เก็บ session ตามข้อกำหนด |
| Refresh token | FR-AUTH-01, NFR-REL-01 | รูปแบบรายละเอียดต้องออกแบบให้สอดคล้องกับ spec |
| หน้าบ้านและหลังบ้าน | ทีมเลือกเอง ไม่ได้มาจาก spec | ยังไม่ระบุเทคโนโลยี |

## 3. โมเดลข้อมูล

| Entity | ฟิลด์หลัก | รองรับ |
|---|---|---|
| Session | session identifier, member reference, refresh token state, expiration state | FR-AUTH-01, NFR-REL-01, CON-CACHE-01 |

รายละเอียดอายุ session ยังไม่สร้างจนกว่าจะได้คำตอบ Q2

## 4. API / หน้าจอ

| รายการ | Input / Output หลัก | รองรับ |
|---|---|---|
| POST `/auth/refresh` | refresh token / session ใหม่ | FR-AUTH-01 |
| DELETE `/auth/session` | session identifier / ผลการออกจาก session | FR-AUTH-01 |

## 5. ตารางตรวจ Constraints

| Constraint ID | ถูกนำไปใช้ที่ไหนใน plan | สถานะ |
|---|---|---|
| CON-CACHE-01 | โมเดล Session และการจัดการ session ด้วย Redis | ใช้แล้ว |

## 6. แผนทดสอบจาก Acceptance Criteria

ยังไม่มี Acceptance Criteria ใน spec จึงยังสร้าง test ที่อ้าง AC ไม่ได้

## 7. ลำดับงาน

1. ยืนยันกลุ่มผู้ใช้และอายุ session ตาม Q2
2. ออกแบบวงจร refresh token ตาม FR-AUTH-01
3. ออกแบบข้อมูล session ใน Redis ตาม CON-CACHE-01
4. กำหนดตัวชี้วัด session หลุดตาม NFR-REL-01
5. สร้าง endpoint ต่ออายุและยกเลิก session ตาม FR-AUTH-01
6. เพิ่ม Acceptance Criteria ก่อนเริ่มทดสอบ

## 8. สิ่งที่ยังไม่ทำ

- Q2 “User 2” หมายถึงผู้ใช้กลุ่มใด และ session ต้องคงอยู่นานเท่าใด
  ส่วนที่เกี่ยวข้องกับข้อนี้จะยังไม่สร้างจนกว่าจะได้คำตอบ

