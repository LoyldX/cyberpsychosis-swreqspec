package roster

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCovers: ตอบว่าคลาสอยู่ในเวลาเข้างานของเทรนเนอร์หรือไม่ ตามกติกาของ 004 และ 005
// (ASM-01, ASM-03, ASM-10; 005 ASM-11, ASM-13).
func TestCovers(t *testing.T) {
	pool := newTestPool(t)
	store := NewStore(pool)
	ctx := context.Background()
	trainer := newTrainer(t, pool, "เอ")
	other := newTrainer(t, pool, "บี")
	noShift := newTrainer(t, pool, "ซี")

	// ศุกร์ (ISO 5): 09:00–12:00 ต่อด้วย 12:00–15:00 และ 16:00–18:00 ที่ถูกยกเลิก
	mustCreate(t, store, trainer, KindWork, slot(5, "09:00", "12:00"))
	mustCreate(t, store, trainer, KindWork, slot(5, "12:00", "15:00"))
	late := mustCreate(t, store, trainer, KindWork, slot(5, "16:00", "18:00"))
	_, err := store.Cancel(ctx, late.ID, changeNow)
	require.NoError(t, err)
	mustCreate(t, store, other, KindWork, slot(5, "18:00", "20:00"))

	cases := []struct {
		name       string
		trainerID  int64
		start, end time.Time
		want       bool
	}{
		{"อยู่ในกะ", trainer, friday(10, 0), friday(11, 0), true},
		{"เริ่มตรงต้นกะ (005 ASM-13)", trainer, friday(9, 0), friday(10, 0), true},
		{"จบตรงสิ้นสุดกะ (005 ASM-13)", trainer, friday(14, 0), friday(15, 0), true},
		{"คร่อมสองกะที่ต่อกัน (ASM-10)", trainer, friday(11, 30), friday(12, 30), true},
		{"เริ่มก่อนกะ (005 AC-BKG-03)", trainer, friday(8, 30), friday(9, 30), false},
		{"จบหลังกะ", trainer, friday(14, 30), friday(15, 30), false},
		{"กะที่ยกเลิกไม่นับ (ASM-01)", trainer, friday(16, 0), friday(17, 0), false},
		{"กะของเทรนเนอร์คนอื่นไม่นับ", trainer, friday(18, 0), friday(19, 0), false},
		{"เทรนเนอร์ไม่มีกะ", noShift, friday(10, 0), friday(11, 0), false},
		{"วันที่ไม่มีกะ (เสาร์)", trainer, friday(10, 0).AddDate(0, 0, 1), friday(11, 0).AddDate(0, 0, 1), false},
		{"ศุกร์สัปดาห์หน้า (ซ้ำทุกสัปดาห์ ASM-03)", trainer, friday(10, 0).AddDate(0, 0, 7), friday(11, 0).AddDate(0, 0, 7), true},
		{"ข้ามวัน", trainer, friday(23, 0), friday(23, 0).Add(2 * time.Hour), false},
		{"เวลา UTC แปลงเป็นเวลาไทยก่อน", trainer, friday(10, 0).UTC(), friday(11, 0).UTC(), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := store.Covers(ctx, tc.trainerID, tc.start, tc.end)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
