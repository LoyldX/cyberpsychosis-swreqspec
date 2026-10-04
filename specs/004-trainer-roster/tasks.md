# Tasks: Trainer Roster
- Feature: Trainer Roster (`004-trainer-roster`)
- Spec ID: FR-ROST-01, CON-DB-01 (AC ทั้งหมดเป็น Candidate)
- อ้างอิง plan.md: `specs/004-trainer-roster/plan.md`
- วันที่: 4 ตุลาคม 2569
- สรุป: ทั้งหมด 5 task, รอ Q-xx 3 task (พร้อมทำ 2 task). ยังไม่มี task ใดถูกเริ่มทำ
- หมายเหตุ: ไฟล์ตามโครงค่าเริ่มต้น (React Vite / FastAPI) ทีมเลือกเอง ไม่ได้มาจาก spec

## รายการ task

### T-01 สร้าง model และ repository interface ของ Trainer roster
- รองรับ: FR-ROST-01, CON-DB-01 (ออกแบบ interface เท่านั้น ยังไม่เลือกฐานข้อมูล)
- ตรวจด้วย: ไม่มี AC ตรง ๆ (งานพื้นฐาน)
- ไฟล์ที่แตะ: `backend/app/models/trainer_roster.py`, `backend/app/repositories/roster_repository.py`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: model มีฟิลด์ trainer reference, work interval, teaching interval และ interface ถูก import ได้
- สถานะ: พร้อมทำ

### T-02 สร้าง GET/PUT ตารางงานเทรนเนอร์ด้วย repository ทดสอบ
- รองรับ: FR-ROST-01
- ตรวจด้วย: AC-ROST-01 (Candidate)
- ไฟล์ที่แตะ: `backend/app/api/roster.py`, `backend/tests/test_trainer_roster.py`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: `test_AC_ROST_01_staff_can_set_and_update_roster` ผ่าน ด้วย repository ทดสอบ (fake)
- สถานะ: พร้อมทำ

### T-03 เชื่อม repository กับฐานข้อมูลจริง
- รองรับ: CON-DB-01, FR-ROST-01
- ตรวจด้วย: AC-ROST-02 (Candidate)
- ไฟล์ที่แตะ: `backend/app/repositories/roster_db.py`, `backend/tests/test_roster_db.py`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: `test_AC_ROST_02_roster_stored_in_relational_db` ผ่าน ด้วยชนิดฐานข้อมูลที่ทีมเลือกตาม Q4
- สถานะ: รอ Q4 (ค่าเดาคือ PostgreSQL แต่ยังไม่เลือก)

### T-04 สร้างหน้าจอ Trainer Roster
- รองรับ: FR-ROST-01 (ข้อเสนอ UI-ROST-01 เป็น Candidate)
- ตรวจด้วย: ไม่มี AC ที่ยอมรับแล้ว (UI-ROST-01 เป็น Candidate)
- ไฟล์ที่แตะ: `frontend/src/pages/TrainerRoster.jsx`, `frontend/src/api/mockRoster.js`
- ต้องทำหลัง: ไม่มี (ใช้ API จำลองตามสัญญาใน plan.md ข้อ 4)
- เสร็จเมื่อ: หน้าจอแสดงและแก้ไขตารางจาก API จำลองได้
- สถานะ: รอ Q-xx (UI-ROST-01 ยังไม่มีหมายเลข Q)

### T-05 ต่อหน้าจอกับ API จริง
- รองรับ: FR-ROST-01, CON-DB-01
- ตรวจด้วย: AC-ROST-01, AC-ROST-02 (Candidate)
- ไฟล์ที่แตะ: `frontend/src/api/roster.js`, `frontend/src/pages/TrainerRoster.jsx`
- ต้องทำหลัง: T-02, T-03, T-04
- เสร็จเมื่อ: บันทึกตารางจากหน้าจอแล้วอ่านกลับได้จากฐานข้อมูลจริง
- สถานะ: รอ Q4 และ Q-xx (UI-ROST-01)

## ตารางตรวจความครบ

### AC ทั้งหมด (Candidate)
| AC ID | task ที่ตรวจ AC นี้ |
|---|---|
| AC-ROST-01 (Candidate) | T-02, T-05 |
| AC-ROST-02 (Candidate) | T-03, T-05 |

### Constraint ทั้งหมด
| Constraint ID | task ที่ทำให้เป็นจริง |
|---|---|
| CON-DB-01 | T-01 (interface), T-03 (เชื่อมฐานข้อมูลจริง), T-05 |

## สิ่งที่ยังไม่ทำ
- **Q4** เลือก PostgreSQL หรือ MySQL: รอ T-03, T-05
- ยังไม่มีหมายเลข Q: UI-ROST-01 (หน้าจอตารางงาน) รอ T-04, T-05
- ยังไม่มีหมายเลข Q: สิทธิ์ละเอียดของเจ้าหน้าที่และผู้จัดการ และรูปแบบตาราง ไม่มี task ในรอบนี้
