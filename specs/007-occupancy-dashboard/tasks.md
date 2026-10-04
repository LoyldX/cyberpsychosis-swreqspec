# Tasks: Occupancy Dashboard
- Feature: Occupancy Dashboard (`007-occupancy-dashboard`)
- Spec ID: FR-BI-01, NFR-PERF-02, DOM-CAP-01, IF-GATE-01 (AC ทั้งหมดเป็น Candidate)
- อ้างอิง plan.md: `specs/007-occupancy-dashboard/plan.md`
- วันที่: 4 ตุลาคม 2569
- สรุป: ทั้งหมด 7 task, รอ Q-xx 7 task (พร้อมทำ 0 task). ทุก task ขึ้นกับ Q3 หรือข้อที่ยังไม่มีหมายเลข Q ยังไม่มี task ใดถูกเริ่มทำ
- หมายเหตุ: ไฟล์ตามโครงค่าเริ่มต้น (React Vite / FastAPI) ทีมเลือกเอง ไม่ได้มาจาก spec

## รายการ task

### T-01 สร้างโครงข้อมูล gate event และ occupancy state
- รองรับ: FR-BI-01, DOM-CAP-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ (งานพื้นฐาน ตรวจผ่าน AC-BI-03 ใน T-03)
- ไฟล์ที่แตะ: `backend/app/models/gate_event.py`, `backend/app/models/occupancy_state.py`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: model มี event reference, direction, event time, current count, last processed event
- สถานะ: รอ Q3 (ฟิลด์ direction และขาออกยังไม่มีคำตอบ)

### T-02 รับ event จากประตูผ่าน POST
- รองรับ: IF-GATE-01, DOM-CAP-01
- ตรวจด้วย: AC-BI-04 (Candidate)
- ไฟล์ที่แตะ: `backend/app/api/gate_events.py`, `backend/tests/test_occupancy.py`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: `test_AC_BI_04_gate_signal_processed_into_count` ผ่าน (โปรโตคอลและรูปแบบสัญญาณเป็น TBD ไม่มี Q)
- สถานะ: รอ Q3 (ทิศทางของ event) และรูปแบบสัญญาณที่ยังไม่มีหมายเลข Q

### T-03 คำนวณจำนวนผู้ใช้บริการตาม DOM-CAP-01
- รองรับ: DOM-CAP-01, FR-BI-01
- ตรวจด้วย: AC-BI-03 (Candidate)
- ไฟล์ที่แตะ: `backend/app/services/occupancy_calc.py`, `backend/tests/test_occupancy.py`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: `test_AC_BI_03_entry_plus_one_exit_minus_one` ผ่าน
- สถานะ: รอ Q3 (ทิศทางขาออก)

### T-04 สร้าง GET จำนวนผู้ใช้บริการปัจจุบัน
- รองรับ: FR-BI-01
- ตรวจด้วย: AC-BI-01 (Candidate)
- ไฟล์ที่แตะ: `backend/app/api/dashboard.py`, `backend/tests/test_occupancy.py`
- ต้องทำหลัง: T-03
- เสร็จเมื่อ: `test_AC_BI_01_dashboard_shows_current_occupancy` ผ่าน
- สถานะ: รอ Q3

### T-05 ทำให้ตัวเลขอัปเดตอัตโนมัติ
- รองรับ: NFR-PERF-02
- ตรวจด้วย: AC-BI-02 (Candidate)
- ไฟล์ที่แตะ: `backend/app/services/occupancy_push.py`, `backend/tests/test_occupancy_refresh.py`
- ต้องทำหลัง: T-04
- เสร็จเมื่อ: `test_AC_BI_02_dashboard_updates_automatically` ผ่านที่รอบเวลาที่ทีมกำหนด
- สถานะ: รอ Q-xx (รอบเวลาอัปเดตของ NFR-PERF-02 ยังไม่มีหมายเลข Q)

### T-06 สร้างหน้า Dashboard
- รองรับ: FR-BI-01 (ข้อเสนอ UI-BI-01/02 เป็น Candidate)
- ตรวจด้วย: ไม่มี AC ที่ยอมรับแล้ว (UI-BI-01/02 เป็น Candidate)
- ไฟล์ที่แตะ: `frontend/src/pages/OccupancyDashboard.jsx`, `frontend/src/api/mockOccupancy.js`
- ต้องทำหลัง: ไม่มี (ใช้ API จำลองตามสัญญาใน plan.md ข้อ 4)
- เสร็จเมื่อ: หน้าจอแสดงจำนวนปัจจุบันเป็นข้อมูลหลักจาก API จำลอง
- สถานะ: รอ Q-xx (UI-BI-01/02 ยังไม่มีหมายเลข Q)

### T-07 ต่อ Dashboard กับ API จริง
- รองรับ: FR-BI-01, IF-GATE-01
- ตรวจด้วย: AC-BI-01, AC-BI-02, AC-BI-04 (Candidate)
- ไฟล์ที่แตะ: `frontend/src/api/occupancy.js`, `frontend/src/pages/OccupancyDashboard.jsx`
- ต้องทำหลัง: T-04, T-05, T-06
- เสร็จเมื่อ: Dashboard แสดงจำนวนจาก API จริงและอัปเดตเอง
- สถานะ: รอ Q3 และ Q-xx (UI-BI-01/02, รอบเวลาอัปเดต)

## ตารางตรวจความครบ

### AC ทั้งหมด (Candidate)
| AC ID | task ที่ตรวจ AC นี้ |
|---|---|
| AC-BI-01 (Candidate) | T-04, T-07 |
| AC-BI-02 (Candidate) | T-05, T-07 |
| AC-BI-03 (Candidate) | T-03 |
| AC-BI-04 (Candidate) | T-02, T-07 |

### Constraint ทั้งหมด
| Constraint ID | task ที่ทำให้เป็นจริง |
|---|---|
| IF-GATE-01 | T-02, T-07 |
| DOM-CAP-01 | T-01, T-02, T-03 |

## สิ่งที่ยังไม่ทำ
- **Q3** ประตูทางเข้ายิมบันทึกทั้งขาเข้าและขาออกหรือไม่: รอ T-01, T-03, T-04, T-07
- **Q10** ปัญหาเครื่องเล่น (อยู่นอกขอบเขตตามค่าเริ่มต้น): ไม่มี task
- ยังไม่มีหมายเลข Q: รอบเวลาอัปเดตของ NFR-PERF-02 รอ T-05
- ยังไม่มีหมายเลข Q: รูปแบบสัญญาณและโปรโตคอลของ IF-GATE-01 รอ T-02
- ยังไม่มีหมายเลข Q: UI-BI-01/02 รอ T-06
