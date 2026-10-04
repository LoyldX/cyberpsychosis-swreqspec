# Tasks: LINE Bot
- Feature: LINE Bot (`003-line-bot`)
- Spec ID: FR-LINE-01, FR-LINE-02, IF-LINE-01 (AC ทั้งหมดเป็น Candidate)
- อ้างอิง plan.md: `specs/003-line-bot/plan.md`
- วันที่: 4 ตุลาคม 2569
- สรุป: ทั้งหมด 8 task, รอ Q-xx 5 task (พร้อมทำ 3 task). ยังไม่มี task ใดถูกเริ่มทำ
- หมายเหตุ: ไฟล์ตามโครงค่าเริ่มต้น (React Vite / FastAPI) ทีมเลือกเอง ไม่ได้มาจาก spec

## รายการ task

### T-01 สร้างโครงข้อมูล FAQ entry
- รองรับ: FR-LINE-02
- ตรวจด้วย: ไม่มี AC ตรง ๆ (งานพื้นฐาน) ใช้ AC-LINE-02 ตรวจในรอบ T-03
- ไฟล์ที่แตะ: `backend/app/models/faq.py`, `backend/tests/test_line_bot.py`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: model FAQ entry (question pattern, answer content) บันทึกและอ่านได้ใน test
- สถานะ: พร้อมทำ

### T-02 สร้าง webhook รับและตอบข้อความผ่าน LINE Messaging API
- รองรับ: IF-LINE-01, FR-LINE-02
- ตรวจด้วย: AC-LINE-03 (Candidate)
- ไฟล์ที่แตะ: `backend/app/api/line_webhook.py`, `backend/app/services/line_client.py`, `backend/tests/test_line_bot.py`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: `test_AC_LINE_03_replies_via_line_messaging_api` ผ่าน (กรณี API ไม่ตอบ = TBD ไม่มี Q จึงไม่ทดสอบในรอบนี้)
- สถานะ: พร้อมทำ

### T-03 ตอบ FAQ อัตโนมัติ
- รองรับ: FR-LINE-02
- ตรวจด้วย: AC-LINE-02 (Candidate)
- ไฟล์ที่แตะ: `backend/app/services/faq_matcher.py`, `backend/tests/test_line_bot.py`
- ต้องทำหลัง: T-01, T-02
- เสร็จเมื่อ: `test_AC_LINE_02_answers_common_faq_automatically` ผ่าน
- สถานะ: พร้อมทำ

### T-04 ผูกบัญชี LINE กับสมาชิก
- รองรับ: FR-LINE-01, IF-LINE-01
- ตรวจด้วย: AC-LINE-01 (Candidate)
- ไฟล์ที่แตะ: `backend/app/models/line_link.py`, `backend/app/services/line_link.py`
- ต้องทำหลัง: T-02
- เสร็จเมื่อ: ผูกบัญชีตามวิธีที่ทีมตอบ Q5 แล้ว test ยืนยันว่าเชื่อม LINE user กับสมาชิกได้
- สถานะ: รอ Q5 (วิธีผูกบัญชียังไม่มีคำตอบ)

### T-05 ค้นข้อมูลการจองของสมาชิกที่ผูกแล้ว
- รองรับ: FR-LINE-01
- ตรวจด้วย: AC-LINE-01 (Candidate)
- ไฟล์ที่แตะ: `backend/app/api/line_bookings.py`, `backend/tests/test_line_bot.py`
- ต้องทำหลัง: T-04 และข้อมูลการจองจาก 005 (T-02 หรือ T-04 ของ 005)
- เสร็จเมื่อ: `test_AC_LINE_01_returns_only_own_booking_status` ผ่าน
- สถานะ: รอ Q5

### T-06 สร้างหน้าจอ FAQ
- รองรับ: FR-LINE-02 (ข้อเสนอ UI-LINE-02 เป็น Candidate)
- ตรวจด้วย: ไม่มี AC ที่ยอมรับแล้ว (UI-LINE-02 เป็น Candidate)
- ไฟล์ที่แตะ: `frontend/src/pages/LineFaq.jsx`, `frontend/src/api/mockFaq.js`
- ต้องทำหลัง: T-03 (ใช้ API จำลองตามสัญญาใน plan.md ข้อ 4)
- เสร็จเมื่อ: หน้าจอแสดงรายการ FAQ จาก API จำลอง
- สถานะ: รอ Q-xx (UI-LINE-02 ยังไม่มีหมายเลข Q)

### T-07 สร้างรูปแบบคำตอบการจอง
- รองรับ: FR-LINE-01 (ข้อเสนอ UI-LINE-01 เป็น Candidate)
- ตรวจด้วย: ไม่มี AC ที่ยอมรับแล้ว (UI-LINE-01 เป็น Candidate)
- ไฟล์ที่แตะ: `backend/app/services/booking_reply.py`
- ต้องทำหลัง: T-05
- เสร็จเมื่อ: คำตอบแสดงคลาส วันเวลา และสถานะการจองของสมาชิกคนนั้น (รูปแบบข้อความทีมเลือกเอง)
- สถานะ: รอ Q5 (UI-LINE-01 เกี่ยวข้องกับ Q5)

### T-08 ต่อ LINE Bot กับ API จริงและทดสอบปลายทาง
- รองรับ: IF-LINE-01, FR-LINE-01, FR-LINE-02
- ตรวจด้วย: AC-LINE-01, AC-LINE-02, AC-LINE-03 (Candidate)
- ไฟล์ที่แตะ: `backend/tests/test_line_bot_e2e.py`
- ต้องทำหลัง: T-05, T-06, T-07
- เสร็จเมื่อ: ทดสอบปลายทางผ่านทั้งสามเส้น (การผูกบัญชีตาม Q5)
- สถานะ: รอ Q5

## ตารางตรวจความครบ

### AC ทั้งหมด (Candidate)
| AC ID | task ที่ตรวจ AC นี้ |
|---|---|
| AC-LINE-01 (Candidate) | T-05, T-08 |
| AC-LINE-02 (Candidate) | T-03, T-08 |
| AC-LINE-03 (Candidate) | T-02, T-08 |

### Constraint ทั้งหมด
| Constraint ID | task ที่ทำให้เป็นจริง |
|---|---|
| IF-LINE-01 | T-02, T-04, T-08 |

## สิ่งที่ยังไม่ทำ
- **Q5** ผู้ใช้ยืนยันตัวตนบน LINE Bot อย่างไรเพื่อผูกกับข้อมูลการจองของตน: รอ T-04, T-05, T-07, T-08
- ยังไม่มีหมายเลข Q: UI-LINE-02 (หน้าจอ FAQ) รอ T-06 เพราะยังไม่มีหมายเลข Q
- ยังไม่มีหมายเลข Q: พฤติกรรมเมื่อ LINE Messaging API ไม่ตอบสนอง ไม่มี task ในรอบนี้
