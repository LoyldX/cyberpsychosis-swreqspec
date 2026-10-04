# แผน: LINE Bot
Spec: `specs/003-line-bot/spec.md` (Draft v2 — ยังไม่ผ่าน clarify) | สร้างเมื่อ 4 ตุลาคม 2569 | หน้านี้เป็นผลลัพธ์ที่ AI ร่าง คนตรวจต้องยืนยัน

## 1. สรุปแนวทาง
ระบบรับและตอบข้อความผ่าน LINE Messaging API ตาม IF-LINE-01
สมาชิกเรียกดูข้อมูลและสถานะการจองของตนตาม FR-LINE-01 และระบบตอบคำถามที่พบบ่อยอัตโนมัติตาม FR-LINE-02
การผูกบัญชี LINE กับสมาชิกยังไม่มีวิธีที่ตกลงกัน (Q5) จึงยังไม่สร้างการค้นข้อมูลการจองที่ผูกกับสมาชิกจริง
ข้อมูลการจองมาจาก 005 และจะยังไม่มีค่าจริงจนกว่า 005 มีข้อมูล

## 2. เทคโนโลยีที่ใช้
| สิ่งที่เลือก | มาจาก | หมายเหตุ |
|---|---|---|
| LINE Messaging API | IF-LINE-01 | เป็น interface ที่ spec ระบุ |
| หน้าบ้าน: React (Vite) | ทีมเลือกเอง ไม่ได้มาจาก spec | ใช้เฉพาะหน้าจอ FAQ (UI-LINE-02 Candidate) ถ้าทีมยืนยัน |
| หลังบ้าน: Python FastAPI | ทีมเลือกเอง ไม่ได้มาจาก spec | ค่าเริ่มต้นของรายวิชา |

## 3. โมเดลข้อมูล
| Entity | ฟิลด์หลัก | รองรับ |
|---|---|---|
| LINE account link | LINE user reference, member reference, link status | FR-LINE-01, IF-LINE-01 (ความสัมพันธ์รอ Q5) |
| Booking summary (อ่านจาก 005) | member reference, booking reference, booking status | FR-LINE-01 |
| FAQ entry | question pattern, answer content | FR-LINE-02 |

- ความสัมพันธ์ระหว่าง LINE account กับสมาชิกไม่สร้างจนกว่าจะตอบ Q5

## 4. API / หน้าจอ
| รายการ | Input / Output หลัก | รองรับ |
|---|---|---|
| POST `/line/webhook` | LINE event / reply result | IF-LINE-01, FR-LINE-02 |
| GET `/line/bookings` | LINE identity (ผ่านการผูกบัญชี รอ Q5) / booking status ของสมาชิก | FR-LINE-01 |
| หน้าจอ FAQ (UI-LINE-02, Candidate) | คำถามและคำตอบ / รายการ FAQ | FR-LINE-02 |

## 5. ตารางตรวจ Constraints
| Constraint ID | ถูกนำไปใช้ที่ไหนใน plan | สถานะ |
|---|---|---|
| IF-LINE-01 | POST `/line/webhook` (ข้อ 4) รับและตอบข้อความผ่าน LINE Messaging API | ใช้แล้ว (พฤติกรรมเมื่อ API ไม่ตอบ = TBD ไม่มี Q) |

## 6. แผนทดสอบจาก Acceptance Criteria
ทุก AC เป็น Candidate (ยังไม่ผ่านการยืนยันจากทีม)

| AC ID | ชื่อ test | ทดสอบอย่างไร |
|---|---|---|
| AC-LINE-01 (Candidate) | `test_AC_LINE_01_returns_only_own_booking_status` | จำลองสมาชิกที่ผูกบัญชีแล้วขอดูการจอง ตรวจว่าได้เฉพาะของตน (การผูกบัญชีรอ Q5) |
| AC-LINE-02 (Candidate) | `test_AC_LINE_02_answers_common_faq_automatically` | ส่งคำถาม FAQ ที่กำหนดไว้ ตรวจว่าได้คำตอบอัตโนมัติ |
| AC-LINE-03 (Candidate) | `test_AC_LINE_03_replies_via_line_messaging_api` | จำลอง event ไปที่ webhook ตรวจว่าเรียก LINE Messaging API และส่งคำตอบ (กรณี API ไม่ตอบ = TBD ไม่มี Q) |

## 7. ลำดับงาน
1. ตอบ Q5 เรื่องการยืนยันตัวตนและผูกบัญชี (รอคำตอบทีม)
2. สร้าง FAQ entry และการจับคู่คำถาม (FR-LINE-02, AC-LINE-02)
3. สร้าง POST `/line/webhook` ตาม IF-LINE-01 (AC-LINE-03)
4. กำหนดข้อมูลการจองที่อ่านจาก 005 (FR-LINE-01)
5. สร้าง GET `/line/bookings` หลังผูกบัญชีได้ (AC-LINE-01, รอ Q5)
6. สร้างหน้าจอ FAQ และรูปแบบคำตอบการจอง (UI-LINE-01/02 Candidate, รอ Q5 และทีม)

## 8. สิ่งที่ยังไม่ทำ
- **Q5** ผู้ใช้งานยืนยันตัวตนบน LINE Bot อย่างไรเพื่อเชื่อมโยงกับข้อมูลการจองของตนเอง
  ส่วนที่เกี่ยวข้องกับข้อนี้จะยังไม่สร้างจนกว่าจะได้คำตอบ (ข้อ 3 LINE account link, ข้อ 4 GET `/line/bookings`, AC-LINE-01, UI-LINE-01)
- ยังไม่มีหมายเลข Q: พฤติกรรมเมื่อ LINE Messaging API ไม่ตอบสนอง (IF-LINE-01) ยังไม่สร้างจนกว่าทีมจะกำหนด
