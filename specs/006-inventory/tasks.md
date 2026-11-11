# Tasks: จัดการสินค้าคงคลังและอุปกรณ์

- Feature: จัดการสินค้าคงคลังและอุปกรณ์
- Spec ID: FR-INV-01, FR-INV-02, FR-INV-03, NFR-DATA-01, DOM-INV-01, DOM-INV-02, CON-DB-01, CON-TECH-01, CON-TECH-02
- อ้างอิง: [spec.md](./spec.md), [plan.md](./plan.md)
- วันที่: 10 พฤศจิกายน 2569

มีทั้งหมด 19 tasks แบ่งงานฐานข้อมูล/API, หน้าจอ และการเชื่อมต่อ API จริง
มี 0 tasks ที่รอ Open Questions; Q-01 เป็นเรื่อง authentication/authorization ซึ่งอยู่นอก scope จึงไม่มี task สำหรับเรื่องนี้

## รายการ task

### T-01 สร้าง migration และโมเดล Inventory
- รองรับ: FR-INV-01, FR-INV-02, FR-INV-03, DOM-INV-02, CON-DB-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-02, T-05 และ T-12
- ไฟล์ที่แตะ: `backend/migrations/0002_create_inventory.up.sql`, `backend/migrations/0002_create_inventory.down.sql`, `backend/internal/inventory/model.go`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: migration สร้างและย้อน schema ของ inventory items กับ movements ได้ และรายการรองรับ Reorder Point ที่ไม่บังคับ
- สถานะ: เสร็จ

### T-02 สร้างที่เก็บข้อมูล Inventory
- รองรับ: FR-INV-01, FR-INV-02, CON-DB-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-03 และ T-05
- ไฟล์ที่แตะ: `backend/internal/inventory/store.go`, `backend/internal/inventory/store_test.go`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: store อ่าน/เขียนรายการและ movement ผ่าน PostgreSQL และซ่อนรายการที่ลบจาก catalog โดยไม่ลบประวัติ
- สถานะ: เสร็จ

### T-03 สร้าง API สำหรับจัดการรายการ
- รองรับ: FR-INV-01, CON-TECH-01, CON-DB-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-04
- ไฟล์ที่แตะ: `backend/internal/inventory/handler.go`, `backend/internal/inventory/routes.go`, `backend/internal/app/router.go`, `backend/cmd/api/main.go`
- ต้องทำหลัง: T-02
- เสร็จเมื่อ: API สร้าง อ่าน ค้นหา แก้ไข และซ่อนรายการทำงานผ่าน Gin โดยรายการใหม่มียอดเริ่มต้น 0
- สถานะ: เสร็จ

### T-04 ทดสอบ API รายการตาม AC
- รองรับ: FR-INV-01, CON-TECH-01, CON-DB-01
- ตรวจด้วย: AC-INV-01, AC-INV-02, AC-INV-03, AC-INV-04, AC-INV-05
- ไฟล์ที่แตะ: `backend/internal/inventory/handler_test.go`
- ต้องทำหลัง: T-03
- เสร็จเมื่อ: `TestAC_INV_01` ถึง `TestAC_INV_05` ผ่าน โดยตรวจรายการและยอดจาก PostgreSQL
- สถานะ: เสร็จ

### T-05 สร้างการเบิกและรับเข้าสต็อกแบบ transaction
- รองรับ: FR-INV-02, NFR-DATA-01, DOM-INV-01, CON-DB-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-06, T-09 และ T-10
- ไฟล์ที่แตะ: `backend/internal/inventory/service.go`, `backend/internal/inventory/store.go`, `backend/internal/inventory/handler.go`
- ต้องทำหลัง: T-01, T-02
- เสร็จเมื่อ: การเบิกหรือรับเข้าบันทึก movement และเปลี่ยนยอดใน transaction เดียว โดยยอดคงเหลือไม่ติดลบ
- สถานะ: เสร็จ

### T-06 ทดสอบการเบิกและรับเข้าสต็อก
- รองรับ: FR-INV-02, DOM-INV-01, NFR-DATA-01
- ตรวจด้วย: AC-INV-06, AC-INV-09
- ไฟล์ที่แตะ: `backend/internal/inventory/handler_test.go`
- ต้องทำหลัง: T-05
- เสร็จเมื่อ: `TestAC_INV_06` และ `TestAC_INV_09` ผ่าน โดยตรวจยอดและจำนวนประวัติ movement
- สถานะ: เสร็จ

### T-07 เพิ่มการยืมและคืนอุปกรณ์
- รองรับ: FR-INV-02, DOM-INV-01, NFR-DATA-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-08
- ไฟล์ที่แตะ: `backend/internal/inventory/service.go`, `backend/internal/inventory/handler.go`
- ต้องทำหลัง: T-05
- เสร็จเมื่อ: API ยืมและคืนอุปกรณ์บันทึก movement และปรับยอดตามทิศทางของรายการ
- สถานะ: เสร็จ

### T-08 ทดสอบการยืมและคืนอุปกรณ์
- รองรับ: FR-INV-02, DOM-INV-01
- ตรวจด้วย: AC-INV-07, AC-INV-08
- ไฟล์ที่แตะ: `backend/internal/inventory/handler_test.go`
- ต้องทำหลัง: T-07
- เสร็จเมื่อ: `TestAC_INV_07` และ `TestAC_INV_08` ผ่าน โดยตรวจยอดและประวัติการยืม/คืน
- สถานะ: เสร็จ

### T-09 ตรวจข้อมูลบังคับของรายการเบิกและยืม
- รองรับ: FR-INV-02, DOM-INV-01
- ตรวจด้วย: AC-INV-10, AC-INV-11, AC-INV-12
- ไฟล์ที่แตะ: `backend/internal/inventory/handler.go`, `backend/internal/inventory/handler_test.go`
- ต้องทำหลัง: T-05, T-07
- เสร็จเมื่อ: `TestAC_INV_10`, `TestAC_INV_11` และ `TestAC_INV_12` ผ่าน โดยคำขอที่ขาดข้อมูลถูกปฏิเสธและยอดคงเดิม
- สถานะ: เสร็จ

### T-10 จัดการข้อผิดพลาดจากฐานข้อมูล
- รองรับ: CON-DB-01, NFR-DATA-01, ASM-07
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-11
- ไฟล์ที่แตะ: `backend/internal/inventory/service.go`, `backend/internal/inventory/handler.go`, `backend/internal/inventory/handler_test.go`
- ต้องทำหลัง: T-05
- เสร็จเมื่อ: เมื่อ transaction บันทึกไม่สำเร็จ ระบบส่ง error กลับและ rollback โดยไม่มีการเปลี่ยนยอดหรือประวัติแบบบางส่วน
- สถานะ: เสร็จ

### T-11 ทดสอบคำขอพร้อมกันด้วย PostgreSQL
- รองรับ: NFR-DATA-01, CON-DB-01
- ตรวจด้วย: AC-INV-14
- ไฟล์ที่แตะ: `backend/internal/inventory/concurrency_test.go`
- ต้องทำหลัง: T-05, T-10
- เสร็จเมื่อ: `TestAC_INV_14` ใช้ PostgreSQL จริงและยืนยันว่าคำขอเบิกพร้อมกันสองคำขอสำเร็จหนึ่งคำขอ ปฏิเสธหนึ่งคำขอ และยอดสุดท้ายเป็น 2
- สถานะ: เสร็จ

### T-12 สร้าง API แจ้งเตือนสต็อกขั้นต่ำ
- รองรับ: FR-INV-03, DOM-INV-02, CON-TECH-01
- ตรวจด้วย: ไม่มี AC ตรง ๆ เป็นงานพื้นฐานของ T-13
- ไฟล์ที่แตะ: `backend/internal/inventory/store.go`, `backend/internal/inventory/service.go`, `backend/internal/inventory/handler.go`, `backend/internal/inventory/routes.go`
- ต้องทำหลัง: T-05
- เสร็จเมื่อ: API ส่งคืนรายการที่กำหนด Reorder Point และมียอดถึงหรือต่ำกว่าจุดดังกล่าวเท่านั้น
- สถานะ: เสร็จ

### T-13 ทดสอบการแจ้งเตือนสต็อกขั้นต่ำ
- รองรับ: FR-INV-03, DOM-INV-02
- ตรวจด้วย: AC-INV-13
- ไฟล์ที่แตะ: `backend/internal/inventory/handler_test.go`
- ต้องทำหลัง: T-12
- เสร็จเมื่อ: `TestAC_INV_13` ตรวจว่ารายการ A ที่ลดถึง Reorder Point ปรากฏในผลแจ้งเตือน ส่วนรายการ B ที่ไม่มี Reorder Point ไม่ปรากฏ
- สถานะ: เสร็จ

### T-14 สร้างหน้าจอ Catalog ด้วย API จำลอง
- รองรับ: FR-INV-01, CON-TECH-02
- ตรวจด้วย: AC-INV-01, AC-INV-02, AC-INV-03, AC-INV-04, AC-INV-05
- ไฟล์ที่แตะ: `frontend/src/features/inventory/InventoryCatalogPage.tsx`, `frontend/src/features/inventory/InventoryCatalogPage.test.tsx`, `frontend/src/features/inventory/types.ts`, `frontend/src/App.tsx`
- ต้องทำหลัง: ไม่มี
- เสร็จเมื่อ: หน้าจอแสดง ค้นหา เพิ่ม แก้ไข และซ่อนรายการโดยใช้ API จำลอง และมี test ที่อ้าง AC ครบ
- สถานะ: เสร็จ

### T-15 สร้างฟอร์ม Stock Movement ด้วย API จำลอง
- รองรับ: FR-INV-02, DOM-INV-01, CON-TECH-02
- ตรวจด้วย: AC-INV-06, AC-INV-07, AC-INV-08, AC-INV-09, AC-INV-10, AC-INV-11, AC-INV-12
- ไฟล์ที่แตะ: `frontend/src/features/inventory/StockMovementForm.tsx`, `frontend/src/features/inventory/StockMovementForm.test.tsx`, `frontend/src/features/inventory/types.ts`
- ต้องทำหลัง: T-14
- เสร็จเมื่อ: ฟอร์มส่งข้อมูลเบิก ยืม คืน และรับเข้าโดยใช้ API จำลอง พร้อมทดสอบข้อมูลบังคับตาม AC-INV-10..12
- สถานะ: เสร็จ

### T-16 สร้างหน้าจอ Low Stock Alert ด้วย API จำลอง
- รองรับ: FR-INV-03, DOM-INV-02, CON-TECH-02
- ตรวจด้วย: AC-INV-13
- ไฟล์ที่แตะ: `frontend/src/features/inventory/LowStockAlerts.tsx`, `frontend/src/features/inventory/LowStockAlerts.test.tsx`, `frontend/src/features/inventory/types.ts`
- ต้องทำหลัง: T-15
- เสร็จเมื่อ: หน้าเว็บแสดงรายการที่แจ้งเตือนได้จาก API จำลอง และมี test ว่ารายการไม่มี Reorder Point ไม่แสดง
- สถานะ: เสร็จ

### T-17 เชื่อมหน้าจอ Catalog กับ API จริง
- รองรับ: FR-INV-01, CON-TECH-02, CON-TECH-01
- ตรวจด้วย: AC-INV-01, AC-INV-02, AC-INV-03, AC-INV-04, AC-INV-05
- ไฟล์ที่แตะ: `frontend/src/features/inventory/api.ts`, `frontend/src/features/inventory/InventoryCatalogPage.tsx`, `frontend/src/features/inventory/InventoryCatalogPage.test.tsx`
- ต้องทำหลัง: T-03, T-04, T-14
- เสร็จเมื่อ: หน้าจอ Catalog ใช้ API จริงสำหรับเพิ่ม อ่าน ค้นหา แก้ไข และซ่อนรายการ โดย test ของ AC ที่เกี่ยวข้องยังผ่าน
- สถานะ: เสร็จ

### T-18 เชื่อมฟอร์ม Movement กับ API จริง
- รองรับ: FR-INV-02, DOM-INV-01, NFR-DATA-01, CON-TECH-01, CON-TECH-02
- ตรวจด้วย: AC-INV-06, AC-INV-07, AC-INV-08, AC-INV-09, AC-INV-10, AC-INV-11, AC-INV-12
- ไฟล์ที่แตะ: `frontend/src/features/inventory/api.ts`, `frontend/src/features/inventory/StockMovementForm.tsx`, `frontend/src/features/inventory/StockMovementForm.test.tsx`
- ต้องทำหลัง: T-06, T-08, T-09, T-15, T-17
- เสร็จเมื่อ: ฟอร์ม Movement ใช้ API จริงและกรณีสำเร็จ/ปฏิเสธแสดงผลสอดคล้องกับผล API โดย test ของ AC ที่เกี่ยวข้องผ่าน
- สถานะ: เสร็จ

### T-19 เชื่อมหน้าแจ้งเตือนกับ API จริง
- รองรับ: FR-INV-03, DOM-INV-02, CON-TECH-01, CON-TECH-02
- ตรวจด้วย: AC-INV-13
- ไฟล์ที่แตะ: `frontend/src/features/inventory/api.ts`, `frontend/src/features/inventory/LowStockAlerts.tsx`, `frontend/src/features/inventory/LowStockAlerts.test.tsx`
- ต้องทำหลัง: T-13, T-16, T-18
- เสร็จเมื่อ: หน้า Low Stock Alert แสดงผลจาก API จริงและ `AC-INV-13` ผ่านการตรวจ end-to-end ระดับเว็บ
- สถานะ: เสร็จ

## ตารางตรวจความครบ

### Acceptance Criteria
| AC ID | task ที่ตรวจ AC นี้ |
|---|---|
| AC-INV-01 | T-04, T-14, T-17 |
| AC-INV-02 | T-04, T-14, T-17 |
| AC-INV-03 | T-04, T-14, T-17 |
| AC-INV-04 | T-04, T-14, T-17 |
| AC-INV-05 | T-04, T-14, T-17 |
| AC-INV-06 | T-06, T-15, T-18 |
| AC-INV-07 | T-08, T-15, T-18 |
| AC-INV-08 | T-08, T-15, T-18 |
| AC-INV-09 | T-06, T-15, T-18 |
| AC-INV-10 | T-09, T-15, T-18 |
| AC-INV-11 | T-09, T-15, T-18 |
| AC-INV-12 | T-09, T-15, T-18 |
| AC-INV-13 | T-13, T-16, T-19 |
| AC-INV-14 | T-11 |

### Constraints
| Constraint ID | task ที่ทำให้เป็นจริง |
|---|---|
| CON-DB-01 | T-01, T-02, T-05, T-10, T-11 |
| CON-TECH-01 | T-03, T-05, T-12 |
| CON-TECH-02 | T-14, T-15, T-16, T-17, T-18, T-19 |
| DOM-INV-01 | T-05, T-07, T-09, T-15, T-18 |
| DOM-INV-02 | T-01, T-12, T-13, T-16, T-19 |
| IF | ไม่มี IF ใน spec นี้ |

## สิ่งที่ยังไม่ทำ
- **Q-01:** “วิธีระบุตัวตนเจ้าหน้าที่ก่อนใช้ระบบและความแตกต่างของสิทธิ์เจ้าหน้าที่ยังเป็นอย่างไร” (Product Owner) — ไม่มี task รอ Q-01 เพราะ authentication/authorization อยู่นอก scope ของฟีเจอร์นี้; จะไม่สร้างส่วนดังกล่าวจนกว่าจะมีข้อกำหนดและมีการเปลี่ยน scope
