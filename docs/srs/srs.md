# SRS ระบบบริหารจัดการฟิตเนส

แหล่งข้อมูล: `draft/refine-req.md` (Refined v4)  
สถานะ: Draft เพื่อรอทีมตอบ Open Questions  
วันที่: 20 กันยายน 2569

## 1. ขอบเขตระบบ

ระบบครอบคลุมการใช้งานของสมาชิก เจ้าหน้าที่ ผู้จัดการ และเทรนเนอร์ ได้แก่ การรักษา session การแจ้งเตือน LINE Bot การจัดตารางเทรนเนอร์ การจองคลาส การจัดการสินค้าคงคลัง และ dashboard จำนวนผู้ใช้บริการปัจจุบัน

### 1.1 In scope

- FR-AUTH-01, FR-NOTI-01, FR-LINE-01, FR-LINE-02
- FR-ROST-01, FR-BKG-01, FR-BKG-02
- FR-INV-01, FR-INV-02, FR-INV-03, FR-BI-01
- NFR-SEC-01, NFR-PERF-01, NFR-PERF-02, NFR-REL-01, NFR-DATA-01
- DOM-BKG-01, DOM-CAP-01, DOM-INV-01, DOM-INV-02

### 1.2 Out of scope

- POS และการขายสินค้าหน้าร้าน
- เครื่องสแกนเวลาและการเชื่อมต่อ HR
- Data Privacy & Encryption และ PDPA Compliance ตามรายการที่ถูกตัดออกในเอกสารต้นทาง
- FR-NUT-01 ฟังก์ชันคำนวณโภชนาการและสารอาหาร
- IF-NUT-01 การเชื่อมต่อ Nutrition API ภายนอก

## 2. Functional Requirements

| ID | ข้อกำหนด |
|---|---|
| FR-AUTH-01 | ระบบรักษา session login ของสมาชิกแบบ persistent ด้วย refresh token |
| FR-NOTI-01 | ระบบส่ง push notification เรื่องเวลาเปิด-ปิดยิมและข่าวสารประชาสัมพันธ์ |
| FR-LINE-01 | สมาชิกเรียกดูข้อมูลและสถานะการจองคลาสของตนผ่าน LINE Bot ได้ |
| FR-LINE-02 | LINE Bot ตอบคำถามทั่วไปเกี่ยวกับยิมและการใช้งานได้อัตโนมัติ |
| FR-ROST-01 | เจ้าหน้าที่หรือผู้จัดการจัด กำหนด และปรับปรุงตารางงานเทรนเนอร์ได้ |
| FR-BKG-01 | สมาชิกเลือกและจองคลาสที่เปิดรับผ่านระบบออนไลน์ได้ |
| FR-BKG-02 | ระบบตรวจสอบและบล็อกการจองซ้อนเมื่อเทรนเนอร์ไม่ว่างหรือนอกเวลาปฏิบัติงาน |
| FR-INV-01 | เจ้าหน้าที่เพิ่ม แก้ไข ลบ ค้นหา และตรวจสอบจำนวนคงเหลือของอุปกรณ์และพัสดุได้ |
| FR-INV-02 | ระบบบันทึกการเบิกใช้ การยืม-คืน และการรับของเข้าสต็อก |
| FR-INV-03 | ระบบแจ้งเตือนเมื่อจำนวนคงเหลือถึง Reorder Point |
| FR-BI-01 | ระบบแสดงจำนวนผู้ใช้บริการภายในยิมจากข้อมูลการผ่านประตูแบบทันที |

## 3. Quality, domain และ integration requirements

| ประเภท | ID | ข้อกำหนด |
|---|---|---|
| NFR | NFR-SEC-01 | แยกเครือข่ายพนักงานและสมาชิกด้วย VLAN |
| NFR | NFR-PERF-01 | การตรวจสอบคิวว่างและบล็อกคิวต้องประมวลผลทันที โดยค่าความหน่วงที่แน่นอนยังต้องกำหนด |
| NFR | NFR-PERF-02 | Dashboard ต้องอัปเดตข้อมูลอัตโนมัติเมื่อมีการผ่านประตู โดยรอบเวลายังต้องกำหนด |
| NFR | NFR-REL-01 | ใช้ Redis ร่วมกับ refresh token และต้องกำหนดเกณฑ์ session หลุดโดยไม่ตั้งใจ |
| NFR | NFR-DATA-01 | การเพิ่มหรือตัดยอดสินค้าคงคลังต้องเป็น ACID transaction |
| DOM | DOM-BKG-01 | ห้ามจองช่วงที่เทรนเนอร์มีคิวอื่นหรืออยู่นอกกะ |
| DOM | DOM-CAP-01 | จำนวนผู้ใช้ = ทางเข้า +1 และทางออก -1 |
| DOM | DOM-INV-01 | การเบิกหรือยืมต้องเก็บผู้รับผิดชอบ วัตถุประสงค์ และกำหนดคืนเมื่อเป็นการยืม |
| DOM | DOM-INV-02 | สิ่งของจำเป็นต้องมีระดับแจ้งเตือนขั้นต่ำ |
| CON | CON-DB-01 | ใช้ PostgreSQL หรือ MySQL สำหรับข้อมูลระบบและ inventory |
| CON | CON-CACHE-01 | ใช้ Redis สำหรับ session login |
| IF | IF-GATE-01 | เชื่อมต่อข้อมูลการผ่านเข้า-ออกจากประตูยิม |
| IF | IF-LINE-01 | เชื่อมต่อ LINE Messaging API |

## 4. Traceability ไปยัง feature specs

| Feature spec | Requirements |
|---|---|
| `001-auth-session` | FR-AUTH-01, NFR-REL-01, CON-CACHE-01 |
| `002-notification` | FR-NOTI-01 |
| `003-line-bot` | FR-LINE-01, FR-LINE-02, IF-LINE-01 |
| `004-trainer-roster` | FR-ROST-01 |
| `005-class-booking` | FR-BKG-01, FR-BKG-02, NFR-PERF-01, DOM-BKG-01 |
| `006-inventory` | FR-INV-01, FR-INV-02, FR-INV-03, NFR-DATA-01, DOM-INV-01, DOM-INV-02, CON-DB-01 |
| `007-occupancy-dashboard` | FR-BI-01, NFR-PERF-02, DOM-CAP-01, IF-GATE-01 |
| `008-network-segmentation` | NFR-SEC-01 |

## 5. Open Questions

การตัดสินใจต่อไปนี้ยังไม่ถูกนำไปเป็นข้อกำหนดเชิงพฤติกรรมใน feature specs:

- Q1 ประเภทพัสดุใน inventory ครอบคลุมอะไรบ้าง
- Q2 “User 2” คือผู้ใช้กลุ่มใด และ session ต้องคงอยู่นานเท่าใด
- Q3 ประตูบันทึกทั้งขาเข้าและขาออกหรือไม่
- Q4 จะเลือก PostgreSQL หรือ MySQL
- Q5 LINE Bot ยืนยันตัวตนและผูกกับบัญชีสมาชิกอย่างไร
