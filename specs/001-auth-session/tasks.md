# Tasks: Auth Session Persistence
- Feature: Auth Session Persistence (`001-auth-session`)
- Spec ID: FR-AUTH-01, NFR-REL-01, CON-CACHE-01 (ยังไม่มี AC รอทีมกำหนด)
- อ้างอิง plan.md: `specs/001-auth-session/plan.md`
- วันที่: 13 ตุลาคม 2569
- สรุป: ทั้งหมด 5 task, พร้อมทำ 0, รอ Q อย่างน้อย 1 ข้อ 3, รอการตัดสินใจทีมอย่างน้อย 1 เรื่อง 2
- หมายเหตุ: ทุก task ต้องยึดกับ requirement ที่มีใน spec และต้องไม่เพิ่มความต้องการใหม่เกินจาก FR-AUTH-01 / NFR-REL-01 / CON-CACHE-01
- หมายเหตุ: ตรวจด้วยใช้ ID ของ spec แทน AC จนกว่าทีมจะกำหนด AC

## รายการ task

### T-01 สร้างโครงข้อมูล session และ refresh token
- รองรับ: FR-AUTH-01, NFR-REL-01, CON-CACHE-01
- ตรวจด้วย: ยังไม่มี AC รอทีมกำหนด (อ้าง FR-AUTH-01, NFR-REL-01, CON-CACHE-01 แทน)
- ไฟล์ที่แตะ: `backend/src/auth/sessionModel.js` หรือเทียบเท่าใน stack ปัจจุบัน
- ต้องทำหลัง: โครงโปรเจกต์ (ตามโครงงานที่มีอยู่)
- เสร็จเมื่อ: model/session store มี user_id, session_id, refresh_token, expires_at, issued_at และ status ที่ใช้กับ Redis ได้
- สถานะ: รอการตัดสินใจทีม: กลุ่มสมาชิกและเงื่อนไขการเข้าสู่ระบบ; รอ Q: เกณฑ์อัตราการหลุดโดยไม่ตั้งใจ

### T-02 สร้าง endpoint ล็อกอินด้วยรหัสนักศึกษา/SSO และสร้าง refresh token พร้อม TTL 7 วัน
- รองรับ: FR-AUTH-01
- ตรวจด้วย: ยังไม่มี AC รอทีมกำหนด (อ้าง FR-AUTH-01 แทน)
- ไฟล์ที่แตะ: `backend/src/auth/controller.js`, `backend/src/auth/routes.js`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: POST `/auth/login` รับข้อมูลล็อกอินหลักด้วยรหัสนักศึกษา/SSO และคืน access token / refresh token พร้อมค่าหมดอายุ 7 วัน
- สถานะ: รอการตัดสินใจทีม: Schema ของผู้ใช้ที่ใช้นักศึกษา/SSO; รอ Q: การเชื่อม provider ของ SSO

### T-03 สร้าง endpoint refresh token และตรวจสอบ expiry / revoke
- รองรับ: FR-AUTH-01, NFR-REL-01
- ตรวจด้วย: ยังไม่มี AC รอทีมกำหนด (อ้าง FR-AUTH-01, NFR-REL-01 แทน)
- ไฟล์ที่แตะ: `backend/src/auth/controller.js`, `backend/src/auth/routes.js`
- ต้องทำหลัง: T-02
- เสร็จเมื่อ: POST `/auth/refresh` ตรวจว่า refresh token ยังถูกต้องและไม่หมดอายุ และคืน access token ใหม่ตามกติกา
- สถานะ: รอ Q: เกณฑ์อัตราการหลุดโดยไม่ตั้งใจ; รอการตัดสินใจทีม: กลยุทธ์ refresh token rotation

### T-04 เชื่อม Redis สำหรับจัดเก็บ session state และยืนยันค่า TTL
- รองรับ: CON-CACHE-01, NFR-REL-01
- ตรวจด้วย: ยังไม่มี AC รอทีมกำหนด (อ้าง CON-CACHE-01, NFR-REL-01 แทน)
- ไฟล์ที่แตะ: `backend/src/shared/redis.js`, `backend/src/auth/sessionStore.js`
- ต้องทำหลัง: T-01
- เสร็จเมื่อ: Redis เก็บ session state และบันทึกค่า TTL สำหรับ refresh token เป็น 7 วัน ตามที่กำหนด
- สถานะ: รอ Q: เกณฑ์อัตราการหลุดโดยไม่ตั้งใจ

### T-05 สร้าง endpoint logout และยกเลิก token ที่ใช้อยู่
- รองรับ: FR-AUTH-01
- ตรวจด้วย: ยังไม่มี AC รอทีมกำหนด (อ้าง FR-AUTH-01 แทน)
- ไฟล์ที่แตะ: `backend/src/auth/controller.js`, `backend/src/auth/routes.js`
- ต้องทำหลัง: T-02, T-03
- เสร็จเมื่อ: POST `/auth/logout` สามารถยกเลิก session/token ปัจจุบัน และป้องกัน refresh token ที่ถูกยกเลิกจากการใช้ต่อ
- สถานะ: รอการตัดสินใจทีม: กลยุทธ์การบังคับ logout และการรองรับหลายเครื่อง

## ตารางตรวจความครบ

### ID ที่ยอมรับแล้วใน spec
| ID | task ที่รองรับ |
|---|---|
| FR-AUTH-01 | T-01, T-02, T-03, T-05 |
| NFR-REL-01 | T-01, T-03, T-04 |
| CON-CACHE-01 | T-01, T-04 |

### Constraint ทั้งหมด
| Constraint ID | task ที่ทำให้เป็นจริง |
|---|---|
| CON-CACHE-01 | T-04 |

## สิ่งที่ยังไม่ทำ
- ยังไม่มีหมายเลข Q: เกณฑ์อัตราการหลุดโดยไม่ตั้งใจของ NFR-REL-01 ยังไม่กำหนดตัวเลข
- ยังไม่มีหมายเลข Q: การล็อกอินด้วย SSO provider ใดบ้าง และมีเงื่อนไขสมาชิกที่ไม่ใช่นักศึกษาหรือไม่
- ยังไม่มีหมายเลข Q: refresh token rotation หรือ reuse แบบใดเหมาะสม
