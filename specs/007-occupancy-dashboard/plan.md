# แผน: Occupancy Dashboard

## 1. สรุปแนวทาง

ระบบจะรับ event การเข้า-ออกจาก gate ตาม IF-GATE-01
ระบบจะคำนวณจำนวนคนในยิมตาม DOM-CAP-01
Dashboard จะแสดงจำนวนผู้ใช้บริการปัจจุบันตาม FR-BI-01
ข้อมูลจะอัปเดตอัตโนมัติตาม NFR-PERF-02
การรองรับขาเข้าและขาออกจริงยังรอคำตอบ Q3

## 2. เทคโนโลยีที่ใช้

| สิ่งที่เลือก | มาจาก | หมายเหตุ |
|---|---|---|
| Interface รับข้อมูล gate | IF-GATE-01 | รูปแบบสัญญาณ/โปรโตคอลยังไม่ระบุ |
| Dashboard | FR-BI-01 | รูปแบบหน้าจอทีมเลือกเอง ไม่ได้มาจาก spec |
| หน้าบ้านและหลังบ้าน | ทีมเลือกเอง ไม่ได้มาจาก spec | ยังไม่ระบุเทคโนโลยี |

## 3. โมเดลข้อมูล

| Entity | ฟิลด์หลัก | รองรับ |
|---|---|---|
| Gate event | event reference, direction, event time | IF-GATE-01, DOM-CAP-01 |
| Occupancy state | current count, last processed event | FR-BI-01, DOM-CAP-01 |

ฟิลด์ direction และวิธีรับ event ยังรอคำตอบ Q3

## 4. API / หน้าจอ

| รายการ | Input / Output หลัก | รองรับ |
|---|---|---|
| POST `/integrations/gate/events` | event เข้า/ออก / ผลรับ event | IF-GATE-01, DOM-CAP-01 |
| GET `/dashboard/occupancy` | เวลาอ้างอิง / จำนวนผู้ใช้ปัจจุบัน | FR-BI-01 |
| หน้าจอ Occupancy Dashboard | จำนวนปัจจุบันและเวลาปรับปรุงล่าสุด | FR-BI-01, NFR-PERF-02 |

## 5. ตารางตรวจ Constraints

| Constraint ID | ถูกนำไปใช้ที่ไหนใน plan | สถานะ |
|---|---|---|
| IF-GATE-01 | endpoint รับ event จากประตู | ใช้แล้ว แต่รูปแบบ interface ยังไม่กำหนด |
| DOM-CAP-01 | เพิ่ม/ลด occupancy ตาม direction | ใช้แล้ว |

## 6. แผนทดสอบจาก Acceptance Criteria

ยังไม่มี Acceptance Criteria ใน spec จึงยังสร้าง test ที่อ้าง AC ไม่ได้

## 7. ลำดับงาน

1. ตอบ Q3 ว่า gate บันทึกขาเข้าและขาออกหรือไม่
2. ระบุรูปแบบ event ของ IF-GATE-01
3. ออกแบบการคำนวณตาม DOM-CAP-01
4. กำหนดรอบ refresh ของ NFR-PERF-02
5. สร้าง endpoint รับ event และหน้าจอ dashboard ตาม FR-BI-01
6. เพิ่ม Acceptance Criteria และทดสอบ event เข้า/ออก

## 8. สิ่งที่ยังไม่ทำ

- Q3 ประตูทางเข้ายิมบันทึกทั้งขาเข้าและขาออกหรือไม่
  ส่วนที่เกี่ยวข้องกับข้อนี้จะยังไม่สร้างจนกว่าจะได้คำตอบ

