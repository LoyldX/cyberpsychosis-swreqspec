# Tasks: จองคลาสออกกำลังกาย
- Feature: จองคลาสออกกำลังกาย (`005-class-booking`)
- Spec ID: FR-BKG-01, FR-BKG-02, DOM-BKG-01, NFR-PERF-01, CON-DB-01 (AC ทั้งหมดเป็น Candidate)
- อ้างอิง plan.md: `specs/005-class-booking/plan.md`
- วันที่: 4 ตุลาคม 2569
- สรุป: ทั้งหมด 8 task, รอ Q-xx 4 task (พร้อมทำ 4 task). ยังไม่มี task ใดถูกเริ่มทำ
- หมายเหตุ: ไฟล์ตามโครงค่าเริ่มต้น (React Vite / FastAPI) ทีมเลือกเอง ไม่ได้มาจาก spec

## รายการ task

### T-01 สร้าง model Class, Booking และ Trainer availability
- รองรับ: FR-BKG-01, FR-BKG-02, DOM-BKG-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ (งานพื้นฐาน)
- ไฟล์ที่แตะ: `backend/app/models/class_booking.py`, `backend/app/models/trainer_availability.py`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: model ครบฟิลด์ตาม plan.md ข้อ 3 และ import ได้
- สถานะ: พร้อมทำ

### T-02 สร้าง GET คลาสที่เปิดรับ
- รองรับ: FR-BKG-01
- ตรวจด้วย: AC-BKG-01 (Candidate) ในส่วนแสดงคลาสที่เปิดรับ
- ไฟล์ที่แตะ: `backend/app/api/classes.py`, `backend/tests/test_class_booking.py`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: GET `/classes?status=open` คืนเฉพาะคลาสที่เปิดรับ (test ตั้งชื่อ `test_AC_BKG_01_...`)
- สถานะ: พร้อมทำ

### T-03 ตรวจการจองซ้อนและเวลานอกกะก่อนบันทึก
- รองรับ: FR-BKG-02, DOM-BKG-01
- ตรวจด้วย: AC-BKG-02, AC-BKG-03 (Candidate)
- ไฟล์ที่แตะ: `backend/app/services/conflict_check.py`, `backend/tests/test_class_booking.py`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: `test_AC_BKG_02_blocks_overlap_when_trainer_busy_or_off_shift` และ `test_AC_BKG_03_member_cannot_book_conflicting_slot` ผ่าน
- สถานะ: พร้อมทำ

### T-04 สร้าง POST การจองคลาส
- รองรับ: FR-BKG-01, FR-BKG-02, DOM-BKG-01
- ตรวจด้วย: AC-BKG-01 (Candidate)
- ไฟล์ที่แตะ: `backend/app/api/bookings.py`, `backend/tests/test_class_booking.py`
- ต้องทำหลัง: T-02, T-03
- เสร็จเมื่อ: `test_AC_BKG_01_member_books_open_class` ผ่าน และคำขอที่ถูกบล็อกไม่ถูกบันทึก
- สถานะ: พร้อมทำ

### T-05 ทดสอบการจองพร้อมกันและเวลาตอบสนอง
- รองรับ: NFR-PERF-01
- ตรวจด้วย: AC-BKG-04 (Candidate)
- ไฟล์ที่แตะ: `backend/tests/test_booking_concurrency.py`
- ต้องทำหลัง: T-04
- เสร็จเมื่อ: `test_AC_BKG_04_concurrent_requests_no_double_booking` ผ่าน (latency เป็น TBD ยังไม่มีหมายเลข Q)
- สถานะ: รอ Q-xx (latency ของ NFR-PERF-01 ยังไม่มีหมายเลข Q)

### T-06 เชื่อม Booking กับฐานข้อมูลจริง
- รองรับ: CON-DB-01
- ตรวจด้วย: AC-BKG-05 (Candidate)
- ไฟล์ที่แตะ: `backend/app/repositories/booking_db.py`, `backend/tests/test_booking_db.py`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: `test_AC_BKG_05_booking_stored_in_relational_db` ผ่าน ด้วยชนิดฐานข้อมูลที่ทีมเลือกตาม Q4
- สถานะ: รอ Q4

### T-07 สร้างหน้าจอจองคลาสและเหตุผลที่ถูกบล็อก
- รองรับ: FR-BKG-01, FR-BKG-02 (ข้อเสนอ UI-BKG-01/02 เป็น Candidate)
- ตรวจด้วย: ไม่มี AC ที่ยอมรับแล้ว (UI-BKG-01/02 เป็น Candidate)
- ไฟล์ที่แตะ: `frontend/src/pages/ClassBooking.jsx`, `frontend/src/api/mockBooking.js`
- ต้องทำหลัง: ไม่มี (ใช้ API จำลองตามสัญญาใน plan.md ข้อ 4)
- เสร็จเมื่อ: หน้าจอแสดงคลาสที่เปิดรับและข้อความ "จองไม่สำเร็จเพราะเทรนเนอร์ไม่ว่าง" จาก API จำลอง
- สถานะ: รอ Q-xx (UI-BKG-01/02 ยังไม่มีหมายเลข Q)

### T-08 ต่อหน้าจอกับ API จริง
- รองรับ: FR-BKG-01, FR-BKG-02, CON-DB-01
- ตรวจด้วย: AC-BKG-01, AC-BKG-05 (Candidate)
- ไฟล์ที่แตะ: `frontend/src/api/booking.js`, `frontend/src/pages/ClassBooking.jsx`
- ต้องทำหลัง: T-04, T-06, T-07
- เสร็จเมื่อ: จองคลาสจากหน้าจอแล้วบันทึกในฐานข้อมูลจริง
- สถานะ: รอ Q4 และ Q-xx (UI-BKG-01/02)

## ตารางตรวจความครบ

### AC ทั้งหมด (Candidate)
| AC ID | task ที่ตรวจ AC นี้ |
|---|---|
| AC-BKG-01 (Candidate) | T-02, T-04, T-08 |
| AC-BKG-02 (Candidate) | T-03 |
| AC-BKG-03 (Candidate) | T-03 |
| AC-BKG-04 (Candidate) | T-05 |
| AC-BKG-05 (Candidate) | T-06, T-08 |

### Constraint ทั้งหมด
| Constraint ID | task ที่ทำให้เป็นจริง |
|---|---|
| CON-DB-01 | T-01 (model), T-06 (เชื่อมฐานข้อมูลจริง), T-08 |
| DOM-BKG-01 | T-03, T-04 |

## สิ่งที่ยังไม่ทำ
- **Q4** เลือก PostgreSQL หรือ MySQL: รอ T-06, T-08
- ยังไม่มีหมายเลข Q: latency ของ NFR-PERF-01 รอ T-05
- ยังไม่มีหมายเลข Q: UI-BKG-01/02 รอ T-07, T-08
