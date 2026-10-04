# Tasks: Push Notification
- Feature: Push Notification (`002-notification`)
- Spec ID: FR-NOTI-01 (AC ทั้งหมดเป็น Candidate)
- อ้างอิง plan.md: `specs/002-notification/plan.md`
- วันที่: 4 ตุลาคม 2569
- สรุป: ทั้งหมด 5 task, รอ Q-xx 4 task (พร้อมทำ 1 task). ยังไม่มี task ใดถูกเริ่มทำ
- หมายเหตุ: ไฟล์ตามโครงค่าเริ่มต้น (React Vite / FastAPI) ทีมเลือกเอง ไม่ได้มาจาก spec

## รายการ task

### T-01 สร้างโครงข้อมูลข้อความแจ้งเตือน
- รองรับ: FR-NOTI-01
- ตรวจด้วย: AC-NOTI-01 (Candidate)
- ไฟล์ที่แตะ: `backend/app/models/notification.py`, `backend/tests/test_notification.py`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: `test_AC_NOTI_01_message_contains_open_close_or_news` ผ่าน (ข้อความมีประเภทเปิด-ปิดยิมหรือข่าวสาร)
- สถานะ: พร้อมทำ

### T-02 สร้าง endpoint ส่งข้อความ
- รองรับ: FR-NOTI-01
- ตรวจด้วย: AC-NOTI-02 (Candidate)
- ไฟล์ที่แตะ: `backend/app/api/notifications.py`, `backend/tests/test_notification.py`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: `test_AC_NOTI_02_delivered_to_member_app` ผ่าน (ผู้รับกลุ่มและความถี่ยังเป็น TBD)
- สถานะ: รอ Q9 (ช่องทางข่าวสารและ Facebook Page ยังไม่มีคำตอบ)

### T-03 เลือกและเชื่อมบริการ push
- รองรับ: FR-NOTI-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ (ยังไม่มี requirement เรื่องบริการ push)
- ไฟล์ที่แตะ: `backend/app/services/push_client.py`
- ต้องทำหลัง: T-02
- เสร็จเมื่อ: ทีมระบุบริการ push ที่ใช้เป็นลายลักษณ์อักษรแล้ว และ client ส่งข้อความทดสอบได้
- สถานะ: รอ Q-xx (ไม่มีหมายเลข Q: ทีมต้องเลือกบริการ push ก่อน)

### T-04 ต่อการส่งกับแอปสมาชิกและทดสอบการถึงผู้รับ
- รองรับ: FR-NOTI-01
- ตรวจด้วย: AC-NOTI-02 (Candidate)
- ไฟล์ที่แตะ: `backend/tests/test_notification_delivery.py`
- ต้องทำหลัง: T-02, T-03
- เสร็จเมื่อ: ข้อความทดสอบถึงแอปสมาชิกที่เป็นผู้รับ (ผู้รับและความถี่ตาม Q9 หรือทีม)
- สถานะ: รอ Q9

### T-05 สร้างหน้าจอประกาศข่าวสารและเวลาเปิด-ปิด
- รองรับ: FR-NOTI-01 (ข้อเสนอ UI-NOTI-01 เป็น Candidate)
- ตรวจด้วย: ไม่มี AC ที่ยอมรับแล้ว (UI-NOTI-01 เป็น Candidate)
- ไฟล์ที่แตะ: `frontend/src/pages/Announcements.jsx`, `frontend/src/api/mockNotifications.js`
- ต้องทำหลัง: T-01, T-02 (ใช้ API จำลองตามสัญญาใน plan.md ข้อ 4)
- เสร็จเมื่อ: หน้าจอแสดงวันและเวลาเปิด-ปิดในข้อความที่ API จำลองส่งมา
- สถานะ: รอ Q9 (UI-NOTI-01 เกี่ยวข้องกับ Q9)

## ตารางตรวจความครบ

### AC ทั้งหมด (Candidate)
| AC ID | task ที่ตรวจ AC นี้ |
|---|---|
| AC-NOTI-01 (Candidate) | T-01 |
| AC-NOTI-02 (Candidate) | T-02, T-04 |

### Constraint ทั้งหมด
| Constraint ID | task ที่ทำให้เป็นจริง |
|---|---|
| ไม่มี CON, DOM หรือ IF ใน spec นี้ | ไม่มีรายการ |

## สิ่งที่ยังไม่ทำ
- **Q9** Push Notification จะมาแทน Facebook Page หรือใช้คู่กัน และต้องโพสต์ไป Facebook Page หรือไม่: รอ T-02, T-04, T-05
- ยังไม่มีหมายเลข Q: ผู้รับกลุ่มใดและรอบเวลาส่ง รอ T-02 และ T-04
- ยังไม่มีหมายเลข Q: การเลือกบริการ push รอ T-03
