# แผน: LINE Bot

## 1. สรุปแนวทาง

ระบบจะรับข้อความผ่าน LINE Messaging API ตาม IF-LINE-01
สมาชิกจะเรียกดูข้อมูลและสถานะการจองของตนตาม FR-LINE-01
ระบบจะตอบคำถามทั่วไปเกี่ยวกับยิมตาม FR-LINE-02
การค้นข้อมูลการจองต้องผูกกับสมาชิกที่ถูกต้อง
วิธีการยืนยันตัวตนยังรอคำตอบ Q5

## 2. เทคโนโลยีที่ใช้

| สิ่งที่เลือก | มาจาก | หมายเหตุ |
|---|---|---|
| LINE Messaging API | IF-LINE-01 | เป็น interface ที่ spec ระบุ |
| หน้าบ้านและหลังบ้าน | ทีมเลือกเอง ไม่ได้มาจาก spec | ยังไม่ระบุเทคโนโลยี |

## 3. โมเดลข้อมูล

| Entity | ฟิลด์หลัก | รองรับ |
|---|---|---|
| LINE account link | LINE user reference, member reference, link status | FR-LINE-01, IF-LINE-01 |
| Booking summary | member reference, booking reference, booking status | FR-LINE-01 |
| FAQ entry | question pattern, answer content | FR-LINE-02 |

ความสัมพันธ์ระหว่าง LINE account กับสมาชิกยังไม่สร้างจนกว่าจะได้คำตอบ Q5

## 4. API / หน้าจอ

| รายการ | Input / Output หลัก | รองรับ |
|---|---|---|
| POST `/line/webhook` | LINE event / reply result | IF-LINE-01, FR-LINE-01, FR-LINE-02 |
| GET `/line/bookings` | LINE identity / booking status ของสมาชิก | FR-LINE-01 |
| หน้าจอ FAQ | คำถามและคำตอบ / รายการ FAQ | FR-LINE-02 |

## 5. ตารางตรวจ Constraints

| Constraint ID | ถูกนำไปใช้ที่ไหนใน plan | สถานะ |
|---|---|---|
| IF-LINE-01 | webhook และการรับส่งข้อความกับ LINE Messaging API | ใช้แล้ว |

## 6. แผนทดสอบจาก Acceptance Criteria

ยังไม่มี Acceptance Criteria ใน spec จึงยังสร้าง test ที่อ้าง AC ไม่ได้

## 7. ลำดับงาน

1. ตอบ Q5 เรื่องการยืนยันตัวตนและการผูกบัญชี
2. ออกแบบ webhook ตาม IF-LINE-01
3. ออกแบบการค้น booking ของสมาชิกตาม FR-LINE-01
4. ออกแบบการจับคู่ FAQ ตาม FR-LINE-02
5. กำหนดกรณี LINE API ไม่ตอบสนองก่อนพัฒนา
6. เพิ่ม Acceptance Criteria และทดสอบการผูกบัญชี/ตอบกลับ

## 8. สิ่งที่ยังไม่ทำ

- Q5 ผู้ใช้งานยืนยันตัวตนบน LINE Bot อย่างไรเพื่อเชื่อมโยงกับข้อมูลการจองของตนเอง
  ส่วนที่เกี่ยวข้องกับข้อนี้จะยังไม่สร้างจนกว่าจะได้คำตอบ

