# แผน: จองคลาสออกกำลังกาย
Spec: `specs/005-class-booking/spec.md` (Draft v2 — ยังไม่ผ่าน clarify) | สร้างเมื่อ 4 ตุลาคม 2569 | หน้านี้เป็นผลลัพธ์ที่ AI ร่าง คนตรวจต้องยืนยัน

## 1. สรุปแนวทาง
ระบบแสดงคลาสที่เปิดรับและรับการจองจากสมาชิกตาม FR-BKG-01
ก่อนบันทึกการจอง ระบบตรวจความว่างของเทรนเนอร์ตาม FR-BKG-02 และกฎ DOM-BKG-01 (เวลาซ้อนและนอกกะ)
ข้อมูลตารางเทรนเนอร์มาจาก 004 และเก็บใน relational DB ตาม CON-DB-01 (ชนิดรอ Q4)
การตอบสนองตาม NFR-PERF-01 ยังวัดไม่ได้เพราะ latency ยังไม่กำหนด จึงยังไม่ตั้งค่าเวลาตอบสนองในโค้ด

## 2. เทคโนโลยีที่ใช้
| สิ่งที่เลือก | มาจาก | หมายเหตุ |
|---|---|---|
| ฐานข้อมูลเชิงสัมพันธ์สำหรับการจอง | CON-DB-01 | ยังไม่เลือก รอ Q4 (ค่าเดาคือ PostgreSQL) |
| หน้าบ้าน: React (Vite) | ทีมเลือกเอง ไม่ได้มาจาก spec | ใช้กับหน้าจอจอง UI-BKG-01/02 (Candidate) |
| หลังบ้าน: Python FastAPI | ทีมเลือกเอง ไม่ได้มาจาก spec | ค่าเริ่มต้นของรายวิชา |

## 3. โมเดลข้อมูล
| Entity | ฟิลด์หลัก | รองรับ |
|---|---|---|
| Class | class reference, trainer reference, open status, time interval | FR-BKG-01 |
| Booking | booking reference, member reference, class reference, status | FR-BKG-01, CON-DB-01 |
| Trainer availability | trainer reference, work interval, teaching interval | FR-BKG-02, DOM-BKG-01 |

## 4. API / หน้าจอ
| รายการ | Input / Output หลัก | รองรับ |
|---|---|---|
| GET `/classes?status=open` | ตัวกรอง / คลาสที่เปิดรับ | FR-BKG-01 |
| POST `/classes/{classId}/bookings` | member reference / ผลการจองหรือเหตุผลที่ปฏิเสธ | FR-BKG-01, FR-BKG-02, DOM-BKG-01 |
| หน้าจอจองคลาส (UI-BKG-01/02, Candidate) | คลาสที่เปิดรับ / สถานะการจองหรือเหตุผลที่ถูกบล็อก | FR-BKG-01, FR-BKG-02 |

## 5. ตารางตรวจ Constraints
| Constraint ID | ถูกนำไปใช้ที่ไหนใน plan | สถานะ |
|---|---|---|
| DOM-BKG-01 | ตรวจเวลาซ้อนและเวลานอกกะก่อนบันทึก booking ใน POST (ข้อ 4) | ใช้แล้ว |
| CON-DB-01 | โมเดล Booking และ Class (ข้อ 3) เก็บใน relational DB | ยังไม่ได้เชื่อมฐานข้อมูลจริง เพราะรอ Q4 |

## 6. แผนทดสอบจาก Acceptance Criteria
ทุก AC เป็น Candidate (ยังไม่ผ่านการยืนยันจากทีม)

| AC ID | ชื่อ test | ทดสอบอย่างไร |
|---|---|---|
| AC-BKG-01 (Candidate) | `test_AC_BKG_01_member_books_open_class` | สมาชิกเลือกคลาสที่เปิดรับและยืนยัน ตรวจว่าการจองถูกบันทึก |
| AC-BKG-02 (Candidate) | `test_AC_BKG_02_blocks_overlap_when_trainer_busy_or_off_shift` | ตั้งเทรนเนอร์ให้มีคิวอื่นหรืออยู่นอกกะ แล้วส่งคำขอจอง ตรวจว่าถูกบล็อกทันที |
| AC-BKG-03 (Candidate) | `test_AC_BKG_03_member_cannot_book_conflicting_slot` | ทดสอบตามกฎ DOM-BKG-01 ว่าจองช่วงที่ซ้อนหรือนอกกะไม่ได้ |
| AC-BKG-04 (Candidate) | `test_AC_BKG_04_concurrent_requests_no_double_booking` | ส่งคำขอพร้อมกันไปยังคลาสเดียวกัน ตรวจว่าไม่มีการจองซ้อน (latency = TBD ไม่มี Q) |
| AC-BKG-05 (Candidate) | `test_AC_BKG_05_booking_stored_in_relational_db` | บันทึกการจองแล้วตรวจในฐานข้อมูลจริง (ชนิดรอ Q4) |

## 7. ลำดับงาน
1. สร้าง model Class, Booking, Trainer availability (FR-BKG-01, CON-DB-01 ที่ยังไม่ระบุชนิด)
2. สร้าง GET `/classes?status=open` (FR-BKG-01, AC-BKG-01)
3. สร้างการตรวจ conflict และเวลานอกกะ (FR-BKG-02, DOM-BKG-01, AC-BKG-02/03)
4. สร้าง POST `/classes/{classId}/bookings` หลังผ่านการตรวจ (AC-BKG-01)
5. ทดสอบการจองพร้อมกัน (NFR-PERF-01, AC-BKG-04, latency รอคำตอบทีม)
6. เชื่อม Booking กับฐานข้อมูลจริง (CON-DB-01, AC-BKG-05, รอ Q4)
7. สร้างหน้าจอจองและการแสดงเหตุผลที่ถูกบล็อก (UI-BKG-01/02 Candidate, รอทีมยืนยัน)
8. ต่อหน้าจอกับ API จริง

## 8. สิ่งที่ยังไม่ทำ
- ไม่มี Open Question ที่ตรงในหัวข้อ 8.2 ของ refine-req
- ยังไม่มีหมายเลข Q: ค่า latency ของ NFR-PERF-01 ส่วนที่เกี่ยวข้อง (ข้อ 7 ลำดับที่ 5, AC-BKG-04) จะยังไม่สร้างจนกว่าทีมจะกำหนด
- **Q4** เลือก PostgreSQL หรือ MySQL: ส่วนเชื่อมฐานข้อมูลจริง (ข้อ 7 ลำดับที่ 6, AC-BKG-05) จะยังไม่สร้างจนกว่าจะได้คำตอบ
