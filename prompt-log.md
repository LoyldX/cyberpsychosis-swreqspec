# Prompt log

## 20 กันยายน 2569 — คำสั่ง `/clarify` และ `/plan`

### Clarify

ตรวจ `specs/001-auth-session/spec.md` ถึง `specs/008-network-segmentation/spec.md` เทียบกับ `draft/refine-req.md` และ AGENTS.md

ประเด็นที่พบร่วมกัน:

- ทุก spec ยังไม่มี Acceptance Criteria จึงยัง trace ไปยัง test ไม่ได้
- Q1 ประเภทสิ่งของใน inventory ยังไม่ตอบ
- Q2 กลุ่มผู้ใช้ “User 2” และอายุ session ยังไม่ตอบ
- Q3 gate บันทึกขาเข้าและขาออกหรือไม่ยังไม่ตอบ
- Q4 ยังไม่เลือก PostgreSQL หรือ MySQL
- Q5 วิธีผูก LINE Bot กับสมาชิกยังไม่ตอบ
- NFR-PERF-01 ยังไม่มีค่า latency ที่ยืนยัน

ยังไม่มีคำตอบจากทีม จึงไม่ได้แก้ `spec.md`

### Plan

- สร้าง `plan.md` ใน feature specs ทั้ง 8 โฟลเดอร์
- ทุก plan ระบุ traceability ไปยัง ID เดิมใน spec
- ไม่มีการสร้าง FR/NFR/DOM/CON/IF/AC ใหม่
- Open Questions ถูกคงไว้ในหัวข้อ “สิ่งที่ยังไม่ทำ” ของแต่ละ plan
- ส่วนที่ไม่มี AC หรือ constraint ใน spec ถูกระบุว่าไม่สามารถวาง test/ตรวจ constraint เพิ่มได้

