# แผน: Trainer Roster
Spec: `specs/004-trainer-roster/spec.md` (Draft v2 — ยังไม่ผ่าน clarify) | สร้างเมื่อ 4 ตุลาคม 2569 | หน้านี้เป็นผลลัพธ์ที่ AI ร่าง คนตรวจต้องยืนยัน

## 1. สรุปแนวทาง
ระบบให้เจ้าหน้าที่หรือผู้จัดการจัดตารางงานเทรนเนอร์ (เวลาเข้างานและเวลาสอน) ตาม FR-ROST-01
ตารางนี้เก็บใน relational database ตาม CON-DB-01 และเป็นข้อมูลที่ 005 ใช้ตรวจการบล็อกคิว
ชนิดฐานข้อมูลยังรอ Q4 จึงออกแบบ repository เป็น interface ก่อน ไม่เชื่อมฐานข้อมูลจริง
สิทธิ์ละเอียดและรูปแบบตารางยังไม่มีใน spec จึงไม่สร้างกฎสิทธิ์เพิ่ม

## 2. เทคโนโลยีที่ใช้
| สิ่งที่เลือก | มาจาก | หมายเหตุ |
|---|---|---|
| ฐานข้อมูลเชิงสัมพันธ์ | CON-DB-01 | ยังไม่เลือก รอ Q4 (ค่าเดาคือ PostgreSQL) |
| หน้าบ้าน: React (Vite) | ทีมเลือกเอง ไม่ได้มาจาก spec | ใช้กับหน้าจอ UI-ROST-01 (Candidate) |
| หลังบ้าน: Python FastAPI | ทีมเลือกเอง ไม่ได้มาจาก spec | ค่าเริ่มต้นของรายวิชา |

## 3. โมเดลข้อมูล
| Entity | ฟิลด์หลัก | รองรับ |
|---|---|---|
| Trainer roster | trainer reference, work interval, teaching interval | FR-ROST-01, CON-DB-01 |

- ชนิดข้อมูลของ interval และรูปแบบตารางไม่กำหนดใน spec (ดูหมายเหตุ "ยังไม่มีหมายเลข Q" ใน spec)

## 4. API / หน้าจอ
| รายการ | Input / Output หลัก | รองรับ |
|---|---|---|
| GET `/trainers/{trainerId}/roster` | trainer reference / ตารางปัจจุบัน | FR-ROST-01 |
| PUT `/trainers/{trainerId}/roster` | ตารางเวลาใหม่ / ตารางที่บันทึก | FR-ROST-01 |
| หน้าจอ Trainer Roster (UI-ROST-01, Candidate) | ตารางงานและเวลาสอน / ผลการบันทึก | FR-ROST-01 |

## 5. ตารางตรวจ Constraints
| Constraint ID | ถูกนำไปใช้ที่ไหนใน plan | สถานะ |
|---|---|---|
| CON-DB-01 | โมเดล Trainer roster (ข้อ 3) เก็บใน relational DB ผ่าน repository | ยังไม่ได้เชื่อมฐานข้อมูลจริง เพราะรอ Q4 (ใช้แล้วเชิงออกแบบ interface) |

## 6. แผนทดสอบจาก Acceptance Criteria
ทุก AC เป็น Candidate (ยังไม่ผ่านการยืนยันจากทีม)

| AC ID | ชื่อ test | ทดสอบอย่างไร |
|---|---|---|
| AC-ROST-01 (Candidate) | `test_AC_ROST_01_staff_can_set_and_update_roster` | ใช้บัญชีเจ้าหน้าที่/ผู้จัดการ กำหนดและแก้ตาราง แล้วอ่านกลับได้ตรงที่บันทึก (ใช้ repository ทดสอบ) |
| AC-ROST-02 (Candidate) | `test_AC_ROST_02_roster_stored_in_relational_db` | บันทึกตารางในฐานข้อมูลจริง แล้วตรวจว่าอยู่ใน relational DB (ชนิดรอ Q4) |

## 7. ลำดับงาน
1. สร้าง model และ repository interface ของ Trainer roster (FR-ROST-01)
2. สร้าง GET/PUT ตารางด้วย repository ทดสอบ พร้อมตรวจ AC-ROST-01 (FR-ROST-01)
3. เชื่อม repository กับฐานข้อมูลจริงตาม CON-DB-01 (AC-ROST-02, รอ Q4)
4. สร้างหน้าจอ Trainer Roster ด้วย API จำลองตามข้อ 4 (UI-ROST-01 Candidate, รอทีมยืนยัน)
5. ต่อหน้าจอกับ API จริง (รอ Q4 และ UI-ROST-01)

## 8. สิ่งที่ยังไม่ทำ
- ไม่มี Open Question เฉพาะฟีเจอร์นี้ในหัวข้อ 8.2 ของ refine-req
- **Q4** เลือก PostgreSQL หรือ MySQL: ส่วนที่เชื่อมฐานข้อมูลจริง (ข้อ 7 ลำดับที่ 3 และ 5) จะยังไม่สร้างจนกว่าจะได้คำตอบ
- ยังไม่มีหมายเลข Q: สิทธิ์ละเอียดของเจ้าหน้าที่และผู้จัดการ และรูปแบบตาราง ยังไม่สร้างจนกว่าทีมจะกำหนด
