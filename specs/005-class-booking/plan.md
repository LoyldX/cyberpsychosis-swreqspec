# แผน: จองคลาสออกกำลังกาย

## 1. สรุปแนวทาง

ระบบจะแสดงคลาสที่เปิดรับและรับการจองจากสมาชิกตาม FR-BKG-01
ก่อนยืนยันการจอง ระบบจะตรวจสอบความว่างของเทรนเนอร์ตาม FR-BKG-02
การตรวจสอบต้องป้องกันเวลาซ้อนและเวลานอกกะตาม DOM-BKG-01
การตอบสนองของการตรวจสอบต้องเป็นไปตาม NFR-PERF-01
ค่า latency ที่แน่นอนยังไม่ถูกกำหนดใน spec

## 2. เทคโนโลยีที่ใช้

| สิ่งที่เลือก | มาจาก | หมายเหตุ |
|---|---|---|
| ระบบจัดเก็บข้อมูลการจอง | ทีมเลือกเอง ไม่ได้มาจาก spec | CON-DB-01 อยู่ใน source แต่ไม่ได้อยู่ใน spec นี้ |
| หน้าบ้านและหลังบ้าน | ทีมเลือกเอง ไม่ได้มาจาก spec | ยังไม่ระบุเทคโนโลยี |

## 3. โมเดลข้อมูล

| Entity | ฟิลด์หลัก | รองรับ |
|---|---|---|
| Class | class reference, trainer reference, open status, time interval | FR-BKG-01 |
| Booking | booking reference, member reference, class reference, status | FR-BKG-01 |
| Trainer availability | trainer reference, work interval, teaching interval | FR-BKG-02, DOM-BKG-01 |

## 4. API / หน้าจอ

| รายการ | Input / Output หลัก | รองรับ |
|---|---|---|
| GET `/classes?status=open` | ตัวกรอง / คลาสที่เปิดรับ | FR-BKG-01 |
| POST `/classes/{classId}/bookings` | member reference / ผลการจองหรือเหตุผลที่ปฏิเสธ | FR-BKG-01, FR-BKG-02, DOM-BKG-01 |
| หน้าจอจองคลาส | คลาสที่เปิดรับ / สถานะการจอง | FR-BKG-01 |

## 5. ตารางตรวจ Constraints

| Constraint ID | ถูกนำไปใช้ที่ไหนใน plan | สถานะ |
|---|---|---|
| DOM-BKG-01 | ตรวจเวลาซ้อนและเวลานอกกะก่อนบันทึก booking | ใช้แล้ว |

## 6. แผนทดสอบจาก Acceptance Criteria

ยังไม่มี Acceptance Criteria ใน spec จึงยังสร้าง test ที่อ้าง AC ไม่ได้

## 7. ลำดับงาน

1. กำหนดค่า latency ของ NFR-PERF-01
2. ออกแบบข้อมูลคลาสและสถานะเปิดรับตาม FR-BKG-01
3. ออกแบบข้อมูล availability ตาม DOM-BKG-01
4. ออกแบบการตรวจ conflict แบบใช้งานพร้อมกันตาม FR-BKG-02
5. สร้างการบันทึก booking หลังผ่านการตรวจสอบ
6. เพิ่ม Acceptance Criteria และทดสอบกรณีจองซ้อน

## 8. สิ่งที่ยังไม่ทำ

- ค่า latency ที่แน่นอนของ NFR-PERF-01 ยังไม่กำหนด
  ส่วนที่เกี่ยวข้องกับข้อนี้จะยังไม่สร้างจนกว่าจะได้คำตอบ

