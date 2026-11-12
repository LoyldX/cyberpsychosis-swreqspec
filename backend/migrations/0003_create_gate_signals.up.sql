-- FR-BI-01, DOM-CAP-01, IF-GATE-01 (ASM-02, ASM-04, ASM-05): เก็บสัญญาณผ่านประตูทุกครั้งเป็นประวัติ
-- id เรียงตามลำดับที่ระบบได้รับ (ASM-05); occupancy_after = จำนวนปัจจุบันหลังสัญญาณนี้ ไม่ต่ำกว่า 0 (ASM-04)
-- ไม่มีข้อมูลระบุตัวบุคคล และไม่มี gate_id เพราะจำนวนประตูยังไม่รู้ (Q-01)
CREATE TABLE gate_signals (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    direction       TEXT        NOT NULL CHECK (direction IN ('in', 'out')),
    received_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    occupancy_after INTEGER     NOT NULL CHECK (occupancy_after >= 0)
);

-- ใช้หาแถวล่าสุดของวันปัจจุบัน (ASM-03)
CREATE INDEX gate_signals_received_at_idx ON gate_signals (received_at);
