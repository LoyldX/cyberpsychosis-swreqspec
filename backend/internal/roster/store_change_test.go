package roster

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gymapp/backend/internal/testutil"
)

// changeNow คือ "ตอนนี้" ของ test: ศุกร์ 13 พ.ย. 2026 08:00 เวลาไทย (ศุกร์ = ISO 5).
var changeNow = time.Date(2026, 11, 13, 8, 0, 0, 0, bangkok)

// friday คือเวลา hh:mm ของวันศุกร์ 13 พ.ย. 2026 เวลาไทย.
func friday(h, m int) time.Time { return time.Date(2026, 11, 13, h, m, 0, 0, bangkok) }

// newBookedClass สร้างคลาสของ 005 ให้เทรนเนอร์ และถ้า booked ก็สร้างสมาชิกกับการจอง 1 รายการ
// ลบการจองและคลาสก่อนลบสมาชิกตอนจบ test (FK ของ 005).
func newBookedClass(t *testing.T, pool *pgxpool.Pool, trainerID int64, start, end time.Time, booked bool) {
	t.Helper()
	ctx := context.Background()
	var classID int64
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO classes (trainer_id, starts_at, ends_at, capacity) VALUES ($1, $2, $3, 10) RETURNING id`,
		trainerID, start, end).Scan(&classID))
	var memberID int64
	if booked {
		memberID = testutil.CreateMember(t, pool, "password-004").ID
		_, err := pool.Exec(ctx, `INSERT INTO bookings (class_id, member_id) VALUES ($1, $2)`, classID, memberID)
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), `DELETE FROM bookings WHERE class_id = $1`, classID)
		require.NoError(t, err)
		_, err = pool.Exec(context.Background(), `DELETE FROM classes WHERE id = $1`, classID)
		require.NoError(t, err)
	})
}

// entriesOf คืนรายการทั้งหมดของเทรนเนอร์ผ่าน store.
func entriesOf(t *testing.T, store *Store, trainerID int64) []Entry {
	t.Helper()
	got, err := store.ListEntries(context.Background(), trainerID)
	require.NoError(t, err)
	return got
}

// TestStoreUpdate: แก้ทีละรายการแล้วเห็นค่าใหม่ และการแก้ที่ผิดกฎไม่บันทึก (ASM-02, ASM-04, ASM-05).
func TestStoreUpdate(t *testing.T) {
	pool := newTestPool(t)
	store := NewStore(pool)
	ctx := context.Background()
	trainer := newTrainer(t, pool, "เอ")
	morning := mustCreate(t, store, trainer, KindWork, slot(1, "09:00", "12:00"))
	afternoon := mustCreate(t, store, trainer, KindWork, slot(1, "13:00", "17:00"))

	updated, err := store.Update(ctx, morning.ID, slot(1, "08:00", "12:00"), changeNow)
	require.NoError(t, err)
	assert.Equal(t, slot(1, "08:00", "12:00"), updated.Slot)
	assert.Equal(t, KindWork, updated.Kind)

	_, err = store.Update(ctx, morning.ID, slot(1, "08:00", "14:00"), changeNow)
	assert.ErrorIs(t, err, ErrOverlap)
	_, err = store.Update(ctx, morning.ID, slot(1, "14:00", "09:00"), changeNow)
	assert.ErrorIs(t, err, ErrInvalidEntry)
	_, err = store.Update(ctx, 999_999, slot(1, "08:00", "09:00"), changeNow)
	assert.ErrorIs(t, err, ErrEntryNotFound)

	// ย้ายเวลาเข้างานไปวันอื่น ทำให้เวลาสอนวันเดิมหลุด → ปฏิเสธ (ASM-04)
	mustCreate(t, store, trainer, KindTeach, slot(1, "14:00", "15:00"))
	_, err = store.Update(ctx, afternoon.ID, slot(2, "13:00", "17:00"), changeNow)
	assert.ErrorIs(t, err, ErrTeachOutsideWork)

	got := entriesOf(t, store, trainer)
	require.Len(t, got, 3)
	assert.Equal(t, slot(1, "08:00", "12:00"), got[0].Slot)
	assert.Equal(t, slot(1, "13:00", "17:00"), got[1].Slot, "การแก้ที่ถูกปฏิเสธไม่ถูกบันทึก")
}

// TestStoreCancel: ยกเลิกแล้วยังเก็บไว้แต่ไม่นับ ยกเลิกซ้ำได้ผลเดิม และยกเลิกเวลาเข้างานที่มีเวลาสอนอยู่ข้างในไม่ได้ (ASM-01, ASM-04).
func TestStoreCancel(t *testing.T) {
	pool := newTestPool(t)
	store := NewStore(pool)
	ctx := context.Background()
	trainer := newTrainer(t, pool, "เอ")
	monday := mustCreate(t, store, trainer, KindWork, slot(1, "09:00", "12:00"))
	tuesday := mustCreate(t, store, trainer, KindWork, slot(2, "09:00", "12:00"))
	mustCreate(t, store, trainer, KindTeach, slot(2, "10:00", "11:00"))

	cancelled, err := store.Cancel(ctx, monday.ID, changeNow)
	require.NoError(t, err)
	assert.True(t, cancelled.Cancelled)
	again, err := store.Cancel(ctx, monday.ID, changeNow)
	require.NoError(t, err)
	assert.True(t, again.Cancelled)

	_, err = store.Create(ctx, trainer, KindTeach, slot(1, "10:00", "11:00"))
	assert.ErrorIs(t, err, ErrTeachOutsideWork, "เวลาเข้างานที่ยกเลิกไม่นับ")
	mustCreate(t, store, trainer, KindWork, slot(1, "09:00", "12:00")) // ไม่ทับกับรายการที่ยกเลิก

	_, err = store.Cancel(ctx, tuesday.ID, changeNow)
	assert.ErrorIs(t, err, ErrTeachOutsideWork)
	assert.Len(t, entriesOf(t, store, trainer), 4, "รายการที่ยกเลิกยังอยู่ (ASM-01)")
}

// TestStoreDelete: ลบแล้วหายจากระบบ ลบซ้ำได้ไม่พบ และลบเวลาเข้างานที่มีเวลาสอนอยู่ข้างในไม่ได้ (ASM-01, ASM-04).
func TestStoreDelete(t *testing.T) {
	pool := newTestPool(t)
	store := NewStore(pool)
	ctx := context.Background()
	trainer := newTrainer(t, pool, "เอ")
	monday := mustCreate(t, store, trainer, KindWork, slot(1, "09:00", "12:00"))
	tuesday := mustCreate(t, store, trainer, KindWork, slot(2, "09:00", "12:00"))
	teach := mustCreate(t, store, trainer, KindTeach, slot(2, "10:00", "11:00"))

	require.NoError(t, store.Delete(ctx, monday.ID, changeNow))
	assert.ErrorIs(t, store.Delete(ctx, monday.ID, changeNow), ErrEntryNotFound)

	assert.ErrorIs(t, store.Delete(ctx, tuesday.ID, changeNow), ErrTeachOutsideWork)
	require.NoError(t, store.Delete(ctx, teach.ID, changeNow), "ลบเวลาสอนได้เสมอ")
	require.NoError(t, store.Delete(ctx, tuesday.ID, changeNow))
	assert.Empty(t, entriesOf(t, store, trainer))
}

// TestStoreBookedClassGuard: แก้/ยกเลิก/ลบเวลาเข้างานที่ทำให้คลาสที่มีผู้จองและยังไม่เริ่มหลุดกะถูกปฏิเสธ
// ส่วนคลาสที่เริ่มแล้ว คลาสที่ไม่มีผู้จอง และคลาสที่ไม่ได้อยู่ในกะอยู่แล้ว ไม่ขวาง (ASM-07, Q-11).
func TestStoreBookedClassGuard(t *testing.T) {
	pool := newTestPool(t)
	store := NewStore(pool)
	ctx := context.Background()
	trainer := newTrainer(t, pool, "เอ")
	work := mustCreate(t, store, trainer, KindWork, slot(5, "09:00", "12:00"))
	newBookedClass(t, pool, trainer, friday(10, 0), friday(11, 0), true)

	_, err := store.Update(ctx, work.ID, slot(5, "13:00", "17:00"), changeNow)
	assert.ErrorIs(t, err, ErrBookedClassOutsideWork)
	_, err = store.Cancel(ctx, work.ID, changeNow)
	assert.ErrorIs(t, err, ErrBookedClassOutsideWork)
	assert.ErrorIs(t, store.Delete(ctx, work.ID, changeNow), ErrBookedClassOutsideWork)
	assert.Equal(t, []Entry{work}, entriesOf(t, store, trainer), "ไม่บันทึกอะไรเลย")

	// ยังครอบคลาสอยู่ แก้ได้
	_, err = store.Update(ctx, work.ID, slot(5, "10:00", "12:00"), changeNow)
	require.NoError(t, err)

	// คลาสเริ่มแล้ว (ตอนนี้ 10:30) ไม่ขวาง (Q-11)
	_, err = store.Update(ctx, work.ID, slot(5, "13:00", "17:00"), friday(10, 30))
	require.NoError(t, err)
}

// TestStoreBookedClassGuardIgnoresUnbookedAndUncovered: คลาสที่ยังไม่มีผู้จอง หรือไม่ได้อยู่ในกะตั้งแต่ก่อนแก้ ไม่ขวาง (ASM-07).
func TestStoreBookedClassGuardIgnoresUnbookedAndUncovered(t *testing.T) {
	pool := newTestPool(t)
	store := NewStore(pool)
	ctx := context.Background()
	trainer := newTrainer(t, pool, "เอ")
	work := mustCreate(t, store, trainer, KindWork, slot(5, "09:00", "12:00"))
	newBookedClass(t, pool, trainer, friday(10, 0), friday(11, 0), false) // ไม่มีผู้จอง
	newBookedClass(t, pool, trainer, friday(15, 0), friday(16, 0), true)  // มีผู้จองแต่อยู่นอกกะอยู่แล้ว

	_, err := store.Update(ctx, work.ID, slot(5, "09:00", "10:00"), changeNow)
	require.NoError(t, err)
}
