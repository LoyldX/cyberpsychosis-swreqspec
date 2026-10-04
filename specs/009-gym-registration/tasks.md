# Tasks: ลงทะเบียนเข้าใช้ยิม (QR Code / ออนไลน์ล่วงหน้า)
- Feature: ลงทะเบียนเข้าใช้ยิม (`009-gym-registration`)
- Spec ID: FR-REG-01 (Candidate), FR-REG-02 (Candidate), DOM-REG-01 (Candidate), NFR-PERF-03 (Candidate), UI-REG-01..04 (Candidate), CON-DB-01 (accepted, ขึ้นกับ Q8)
- อ้างอิง plan.md: `specs/009-gym-registration/plan.md`
- วันที่: 4 ตุลาคม 2569
- สรุป: ทั้งหมด 8 task, รอ Q-xx 8 task (พร้อมทำ 0 task). ทุก task รอทีมยืนยัน Candidate ยังไม่มี task ใดถูกเริ่มทำ
- หมายเหตุ: ไม่มี requirement ที่ยอมรับแล้วในฟีเจอร์นี้ ทุก task จึงอ้าง CON-DB-01 (accepted) เป็น ID ที่รองรับ และผูกกับ Q8 ตามที่ refine-req ระบุ
- หมายเหตุ: ไฟล์ตามโครงค่าเริ่มต้น (React Vite / FastAPI) ทีมเลือกเอง ไม่ได้มาจาก spec

## รายการ task

### T-01 สร้าง model การลงทะเบียน (รหัสนักศึกษา, ชื่อ-นามสกุล)
- รองรับ: CON-DB-01 (ตาม Q8), DOM-REG-01 (Candidate)
- ตรวจด้วย: AC-REG-03 (Candidate)
- ไฟล์ที่แตะ: `backend/app/models/gym_registration.py`, `backend/tests/test_gym_registration.py`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: model มีเพียงรหัสนักศึกษาและชื่อ-นามสกุลและ registered at (ไม่มีฟิลด์อื่น)
- สถานะ: รอ Q7 และ Q8

### T-02 สร้าง POST การลงทะเบียนและตรวจข้อมูลขั้นต่ำ
- รองรับ: FR-REG-01 (Candidate), DOM-REG-01 (Candidate), CON-DB-01 (ตาม Q8)
- ตรวจด้วย: AC-REG-03 (Candidate)
- ไฟล์ที่แตะ: `backend/app/api/gym_registrations.py`, `backend/tests/test_gym_registration.py`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: `test_AC_REG_03_only_student_id_and_name_collected` ผ่าน
- สถานะ: รอ Q7 และ Q8

### T-03 รองรับการลงทะเบียนออนไลน์ล่วงหน้า
- รองรับ: FR-REG-02 (Candidate), CON-DB-01 (ตาม Q8)
- ตรวจด้วย: AC-REG-02 (Candidate)
- ไฟล์ที่แตะ: `backend/app/services/pre_registration.py`, `backend/tests/test_gym_registration.py`
- ต้องทำหลัง: T-02
- เสร็จเมื่อ: `test_AC_REG_02_online_pre_registration_saved_before_arrival` ผ่าน
- สถานะ: รอ Q6 (ต้องมีทั้งสองช่องทางหรือเลือกหนึ่ง)

### T-04 รองรับการเข้าหน้าลงทะเบียนจาก QR Code
- รองรับ: FR-REG-01 (Candidate), CON-DB-01 (ตาม Q8)
- ตรวจด้วย: AC-REG-01 (Candidate)
- ไฟล์ที่แตะ: `backend/app/services/qr_entry.py`, `backend/tests/test_gym_registration.py`
- ต้องทำหลัง: T-02
- เสร็จเมื่อ: `test_AC_REG_01_qr_scan_opens_registration` ผ่าน
- สถานะ: รอ Q6 (ใครสแกนและทุกครั้งหรือครั้งเดียว)

### T-05 ทดสอบเวลาต่อคนและความยาวคิว
- รองรับ: NFR-PERF-03 (Candidate), CON-DB-01 (ตาม Q8)
- ตรวจด้วย: AC-REG-04 (Candidate)
- ไฟล์ที่แตะ: `backend/tests/test_registration_load.py`
- ต้องทำหลัง: T-03, T-04
- เสร็จเมื่อ: `test_AC_REG_04_registration_time_and_queue_within_limit` ผ่านเกณฑ์ที่ทีมกำหนด
- สถานะ: รอ Q11 (เกณฑ์เป็น TBD)

### T-06 บันทึกการลงทะเบียนในฐานข้อมูลจริง
- รองรับ: CON-DB-01
- ตรวจด้วย: AC-REG-05 (Candidate)
- ไฟล์ที่แตะ: `backend/app/repositories/gym_registration_db.py`, `backend/tests/test_gym_registration_db.py`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: `test_AC_REG_05_registration_stored_in_relational_db` ผ่าน ด้วยชนิดฐานข้อมูลตาม Q4
- สถานะ: รอ Q4 และ Q8

### T-07 สร้างหน้าจอลงทะเบียน 2 ช่องและหน้าสำเร็จ/ไม่สำเร็จ
- รองรับ: CON-DB-01 (ผ่าน API ของ T-02) ข้อเสนอ UI-REG-01..04 เป็น Candidate
- ตรวจด้วย: ไม่มี AC ที่ยอมรับแล้ว (UI-REG-01..04 เป็น Candidate)
- ไฟล์ที่แตะ: `frontend/src/pages/GymRegistration.jsx`, `frontend/src/api/mockRegistration.js`
- ต้องทำหลัง: ไม่มี (ใช้ API จำลองตามสัญญาใน plan.md ข้อ 4)
- เสร็จเมื่อ: หน้าจอมีช่องกรอก 2 ช่องและแสดงผลสำเร็จหรือไม่สำเร็จจาก API จำลอง
- สถานะ: รอ Q6, Q7 และ Q-xx (UI-REG-03/04 ยังไม่มีหมายเลข Q)

### T-08 ต่อหน้าจอลงทะเบียนกับ API จริง
- รองรับ: CON-DB-01 (ตาม Q8)
- ตรวจด้วย: AC-REG-01, AC-REG-02, AC-REG-03 (Candidate)
- ไฟล์ที่แตะ: `frontend/src/api/gymRegistration.js`, `frontend/src/pages/GymRegistration.jsx`
- ต้องทำหลัง: T-02, T-07
- เสร็จเมื่อ: ลงทะเบียนจากหน้าจอแล้วบันทึกผ่าน API จริง
- สถานะ: รอ Q6, Q7

## ตารางตรวจความครบ

### AC ทั้งหมด (Candidate)
| AC ID | task ที่ตรวจ AC นี้ |
|---|---|
| AC-REG-01 (Candidate) | T-04, T-08 |
| AC-REG-02 (Candidate) | T-03, T-08 |
| AC-REG-03 (Candidate) | T-01, T-02, T-08 |
| AC-REG-04 (Candidate) | T-05 |
| AC-REG-05 (Candidate) | T-06 |

### Constraint ทั้งหมด
| Constraint ID | task ที่ทำให้เป็นจริง |
|---|---|
| CON-DB-01 (ขึ้นกับ Q8) | T-01, T-06 (ฐานข้อมูลจริง), T-02 ถึง T-05, T-08 |
| DOM-REG-01 (Candidate) | T-01, T-02 |

## สิ่งที่ยังไม่ทำ
- **Q6** การลงทะเบียนต้องมีทั้ง QR Code และออนไลน์ล่วงหน้าหรือเลือกอย่างเดียว / ใครสแกน / ทุกครั้งหรือครั้งเดียว: รอ T-03, T-04, T-07, T-08
- **Q7** ผู้ใช้บริการมีเฉพาะนักศึกษาหรือไม่ และผู้ไม่มีรหัสนักศึกษาลงทะเบียนอย่างไร: รอ T-01, T-02, T-07, T-08
- **Q8** ข้อมูลลงทะเบียนเกี่ยวข้องกับการนับคนเข้ายิมและ Excel เดิมหรือไม่: รอ T-01 ถึง T-06 (CON-DB-01)
- **Q11** เป้าหมายเวลาลงทะเบียนต่อคนและความยาวคิว: รอ T-05
- **Q12** "Main Prior REQ" คือลำดับความสำคัญสูงสุดหรือไม่: กระทบการจัดลำดับทั้งฟีเจอร์ ไม่มี task เฉพาะ
- **Q4** เลือกชนิดฐานข้อมูล: รอ T-06
- ยังไม่มีหมายเลข Q: UI-REG-03/04 รอ T-07
