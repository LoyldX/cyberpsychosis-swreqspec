# แผน: Occupancy Dashboard
Spec: `specs/007-occupancy-dashboard/spec.md` (Draft v2 — ยังไม่ผ่าน clarify) | สร้างเมื่อ 4 ตุลาคม 2569 | หน้านี้เป็นผลลัพธ์ที่ AI ร่าง คนตรวจต้องยืนยัน

## 1. สรุปแนวทาง
ระบบรับ event การผ่านประตูจาก gate ตาม IF-GATE-01 และคำนวณจำนวนผู้ใช้บริการตาม DOM-CAP-01 (ผ่านเข้า +1 ผ่านออก -1)
Dashboard แสดงจำนวนปัจจุบันตาม FR-BI-01 และอัปเดตอัตโนมัติตาม NFR-PERF-02
การบันทึกขาออกยังรอ Q3 ซึ่งมีผลต่อกฎ DOM-CAP-01 ทั้งหมด จึงยังไม่ทำตรรกะการลดจำนวนจนกว่าจะตอบ
รูปแบบสัญญาณ gate และรอบเวลาอัปเดตยังไม่ระบุใน refine-req

## 2. เทคโนโลยีที่ใช้
| สิ่งที่เลือก | มาจาก | หมายเหตุ |
|---|---|---|
| Interface รับข้อมูล gate | IF-GATE-01 | โปรโตคอลและรูปแบบสัญญาณยังไม่ระบุ (ไม่มีหมายเลข Q) |
| หน้าบ้าน Dashboard: React (Vite) | ทีมเลือกเอง ไม่ได้มาจาก spec | ใช้กับ UI-BI-01/02 (Candidate) |
| หลังบ้าน: Python FastAPI | ทีมเลือกเอง ไม่ได้มาจาก spec | ค่าเริ่มต้นของรายวิชา |

## 3. โมเดลข้อมูล
| Entity | ฟิลด์หลัก | รองรับ |
|---|---|---|
| Gate event | event reference, direction (ขาเข้า / ขาออก), event time | IF-GATE-01, DOM-CAP-01 |
| Occupancy state | current count, last processed event | FR-BI-01, DOM-CAP-01 |

- ฟิลด์ direction จะมีค่าขาออกเมื่อ Q3 ตอบว่าบันทึกขาออก

## 4. API / หน้าจอ
| รายการ | Input / Output หลัก | รองรับ |
|---|---|---|
| POST `/integrations/gate/events` | event เข้า/ออก / ผลรับ event | IF-GATE-01, DOM-CAP-01 |
| GET `/dashboard/occupancy` | เวลาอ้างอิง / จำนวนผู้ใช้บริการปัจจุบัน | FR-BI-01 |
| หน้าจอ Occupancy Dashboard (UI-BI-01/02, Candidate) | จำนวนปัจจุบัน / เวลาปรับปรุงล่าสุด | FR-BI-01, NFR-PERF-02 |

## 5. ตารางตรวจ Constraints
| Constraint ID | ถูกนำไปใช้ที่ไหนใน plan | สถานะ |
|---|---|---|
| IF-GATE-01 | POST `/integrations/gate/events` (ข้อ 4) รับ event จากประตู | ใช้แล้ว (โปรโตคอลยังไม่กำหนด) |
| DOM-CAP-01 | Occupancy state และการคำนวณ +1/-1 (ข้อ 3) | ใช้แล้ว (ทิศทางขาออกรอ Q3) |

## 6. แผนทดสอบจาก Acceptance Criteria
ทุก AC เป็น Candidate (ยังไม่ผ่านการยืนยันจากทีม)

| AC ID | ชื่อ test | ทดสอบอย่างไร |
|---|---|---|
| AC-BI-01 (Candidate) | `test_AC_BI_01_dashboard_shows_current_occupancy` | ป้อน event ชุดหนึ่งแล้วเปิด GET `/dashboard/occupancy` ตรวจว่าตัวเลขตรงกับการคำนวณ (ขึ้นกับ Q3) |
| AC-BI-02 (Candidate) | `test_AC_BI_02_dashboard_updates_automatically` | เปิดหน้าจอ ส่ง event ใหม่ ตรวจว่าตัวเลขอัปเดตเอง (รอบเวลา = TBD ไม่มี Q) |
| AC-BI-03 (Candidate) | `test_AC_BI_03_entry_plus_one_exit_minus_one` | ส่ง event ผ่านเข้า 1 ครั้ง แล้วผ่านออก 1 ครั้ง ตรวจว่า N+1 แล้ว N (ขึ้นกับ Q3) |
| AC-BI-04 (Candidate) | `test_AC_BI_04_gate_signal_processed_into_count` | จำลองสัญญาณประตูผ่าน POST แล้วตรวจว่าเข้าสู่การนับ (โปรโตคอล = TBD ไม่มี Q; ทิศทางรอ Q3) |

## 7. ลำดับงาน
1. ตอบ Q3 ว่า gate บันทึกขาเข้าและขาออกหรือไม่ (รอคำตอบทีม)
2. กำหนดรูปแบบ event ของ IF-GATE-01 (ไม่มี Q รองรับ ต้องให้ทีมตัดสินใจ)
3. สร้าง Gate event และ Occupancy state (FR-BI-01, DOM-CAP-01, รอ Q3 สำหรับ direction)
4. สร้าง POST `/integrations/gate/events` (IF-GATE-01, AC-BI-04)
5. สร้างการคำนวณ +1/-1 (DOM-CAP-01, AC-BI-03, รอ Q3)
6. สร้าง GET `/dashboard/occupancy` และหน้าจอ (FR-BI-01, AC-BI-01, UI-BI-01/02 Candidate)
7. กำหนดรอบเวลาอัปเดตแล้วทดสอบอัตโนมัติ (NFR-PERF-02, AC-BI-02, รอทีมกำหนด)

## 8. สิ่งที่ยังไม่ทำ
- **Q3** ประตูทางเข้ายิมบันทึกทั้งขาเข้าและขาออกหรือไม่
  ส่วนที่เกี่ยวข้องกับข้อนี้จะยังไม่สร้างจนกว่าจะได้คำตอบ (ข้อ 3 direction ขาออก, ข้อ 7 ลำดับที่ 1, 3, 5, AC-BI-01, AC-BI-03)
- **Q10** ปัญหาเครื่องเล่นมีคนใช้อยู่เป็นประจำ: อยู่นอกขอบเขตตามค่าเริ่มต้น ไม่สร้างสิ่งใดเพิ่ม จนกว่าทีมจะตอบ
- ยังไม่มีหมายเลข Q: รอบเวลาอัปเดต (NFR-PERF-02) และรูปแบบสัญญาณ gate (IF-GATE-01) จะยังไม่สร้างจนกว่าทีมจะกำหนด
