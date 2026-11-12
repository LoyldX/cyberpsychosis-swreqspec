package roster

import (
	"context"
	"time"
)

// Covers บอกว่าช่วงเวลา [start, end] (เช่นคลาสของ 005) อยู่ในเวลาเข้างานที่ไม่ถูกยกเลิกของเทรนเนอร์ทั้งหมดหรือไม่
// เทียบตามวันและเวลา Asia/Bangkok ขอบพอดีนับว่าอยู่ใน รวมเวลาเข้างานที่ต่อกันเป็นช่วงเดียว ช่วงที่ข้ามวัน
// หรือเทรนเนอร์ที่ไม่มีเวลาเข้างานวันนั้น = ไม่อยู่ในกะ (FR-ROST-01, CON-DB-01; ASM-01, ASM-03, ASM-10; 005 ASM-11, ASM-13).
func (s *Store) Covers(ctx context.Context, trainerID int64, start, end time.Time) (bool, error) {
	slot, ok := SlotOf(start, end)
	if !ok {
		return false, nil
	}
	entries, err := activeEntries(ctx, s.pool, trainerID, slot.Weekday)
	if err != nil {
		return false, err
	}
	return slot.Within(workSlots(entries)), nil
}
