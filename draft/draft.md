# Infrastructure

## 1. Backend Infrastructure (โครงสร้างระบบคลาวด์และหลังบ้าน)
- **Database Management:**
  - **Relational DB (PostgreSQL / MySQL):** จัดเก็บข้อมูลสมาชิก ตารางงานเทรนเนอร์ การจองคลาส และข้อมูลคลังสินค้า/อุปกรณ์ (Inventory)
  - **In-Memory Database (Redis):** จัดการ Session Login ของแอปพลิเคชัน เพื่อแก้ปัญหา User โดน Auto-logout บ่อย

## 2. Application & Software Ecosystem (ซอฟต์แวร์และแอปพลิเคชัน)
- **Member Mobile App & LINE Messaging API:**
  - ระบบรักษา Session Login แบบ Persistent (ใช้ Refresh Tokens)
  - ระบบ Push Notification แจ้งเตือนวัน/เวลาเปิด-ปิดยิม และข่าวสาร
  - LINE Bot Automation สำหรับดึงข้อมูลการจองและตอบคำถามเบื้องต้นอัตโนมัติ
- **Gym Management & Staff Portal:**
  - ระบบจัดตารางงานเทรนเนอร์ (Trainer Roster) และระบบจองคลาสออนไลน์แบบ Real-time (บล็อกคิวอัตโนมัติเมื่อเทรนเนอร์ไม่ว่าง)
  - ระบบจัดการสินค้าคงคลังและอุปกรณ์ (Inventory Management): บันทึกรายการพัสดุ/อุปกรณ์ ติดตามสถานะสต็อก และการเบิกจ่าย
- **Business Intelligence (BI) Dashboard:**
  - ดึงข้อมูลจากประตูทางเข้า มาแสดงผลจำนวนผู้ใช้บริการแบบ Real-time แทนการคีย์ Excel

## 3. Network Infrastructure (เครือข่าย)
- **Network Segmentation (VLAN):** แยกวงเครือข่าย Wi-Fi ออกจากกันชัดเจน (พนักงาน และ Wi-Fi สมาชิก) เพื่อป้องกันการโจรกรรมข้อมูล
