package booking

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gymapp/backend/internal/testutil"
)

// testNow คือ "ตอนนี้" ของ test ที่ใช้ DB: 08:00 ของวันทดสอบ คลาสใน test ส่วนใหญ่เริ่มหลังจากนี้.
var testNow = at("08:00")

// allowShift คือ ShiftCheck จำลองที่ตอบว่าอยู่ในกะเสมอ ใช้เมื่อ test ไม่ได้ตรวจเรื่องกะ.
func allowShift(context.Context, int64, Interval) error { return nil }

// newTestPool ต่อฐานข้อมูลทดสอบ (migrate ถึง 0004) และล้าง classes/bookings ก่อนและหลัง test
// (plan D8: ฐานข้อมูลเดียวกับ dev; ไม่ใช้ t.Parallel).
func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := testutil.ConnectPool(t)
	resetBookingTables(t, pool)
	t.Cleanup(func() { resetBookingTables(t, pool) })
	return pool
}

// resetBookingTables ลบการจองและคลาสทั้งหมด (ข้อมูลทดสอบของ 005 เท่านั้น).
func resetBookingTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	_, err := pool.Exec(ctx, `DELETE FROM bookings`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `DELETE FROM classes`)
	require.NoError(t, err)
}

// newMember สร้างสมาชิกใน members (001) และลบการจองของสมาชิกนั้นก่อนลบตัวสมาชิกตอนจบ test (FK ของ ASM-06).
func newMember(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()
	m := testutil.CreateMember(t, pool, "password-005")
	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), `DELETE FROM bookings WHERE member_id = $1`, m.ID)
		require.NoError(t, err)
	})
	return m.ID
}

// newClass เพิ่มคลาสที่เตรียมไว้ล่วงหน้า (ASM-04) แล้วคืน id.
func newClass(t *testing.T, pool *pgxpool.Pool, trainerID int64, iv Interval, capacity int) int64 {
	t.Helper()
	var id int64
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO classes (trainer_id, starts_at, ends_at, capacity) VALUES ($1, $2, $3, $4) RETURNING id`,
		trainerID, iv.Start, iv.End, capacity).Scan(&id))
	return id
}

// countBookings นับแถวใน bookings ของคลาส ใช้ยืนยันว่า "ไม่บันทึก" (ASM-05, ASM-07).
func countBookings(t *testing.T, pool *pgxpool.Pool, classID int64) int {
	t.Helper()
	var n int
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT count(*) FROM bookings WHERE class_id = $1`, classID).Scan(&n))
	return n
}

// TestStoreListUpcoming: คืนเฉพาะคลาสที่ยังไม่ถึงเวลาเริ่ม เรียงตามเวลาเริ่ม (FR-BKG-01, plan D4).
func TestStoreListUpcoming(t *testing.T) {
	pool := newTestPool(t)
	store := NewStore(pool)
	ctx := context.Background()

	newClass(t, pool, 1, span("07:00", "08:00"), 10) // เริ่มไปแล้ว
	startsNow := newClass(t, pool, 1, span("08:00", "09:00"), 10)
	later := newClass(t, pool, 2, span("11:00", "12:00"), 5)
	sooner := newClass(t, pool, 1, span("10:00", "11:00"), 10)

	got, err := store.ListUpcoming(ctx, testNow)
	require.NoError(t, err)
	ids := make([]int64, len(got))
	for i, c := range got {
		ids[i] = c.ID
	}
	assert.Equal(t, []int64{sooner, later}, ids, "คลาสที่เริ่ม 08:00 พอดี (id %d) นับว่าเริ่มแล้ว", startsNow)
	assert.Equal(t, int64(2), got[1].TrainerID)
	assert.Equal(t, 5, got[1].Capacity)
	assert.True(t, got[1].StartsAt.Equal(at("11:00")))
}

// TestStoreGetForMember: ดูการจองได้เฉพาะของตนเอง ของคนอื่นหรือไม่มีอยู่ได้ ErrBookingNotFound (plan D5).
func TestStoreGetForMember(t *testing.T) {
	pool := newTestPool(t)
	store := NewStore(pool)
	ctx := context.Background()
	owner, other := newMember(t, pool), newMember(t, pool)
	classID := newClass(t, pool, 1, span("10:00", "11:00"), 10)

	b, err := store.Book(ctx, classID, owner, testNow, allowShift)
	require.NoError(t, err)

	got, err := store.GetForMember(ctx, b.ID, owner)
	require.NoError(t, err)
	assert.Equal(t, b.ID, got.ID)
	assert.Equal(t, classID, got.ClassID)
	assert.Equal(t, owner, got.MemberID)

	_, err = store.GetForMember(ctx, b.ID, other)
	assert.ErrorIs(t, err, ErrBookingNotFound)
	_, err = store.GetForMember(ctx, b.ID+1000, owner)
	assert.ErrorIs(t, err, ErrBookingNotFound)
}

// TestStoreBookRejections: ไม่พบคลาส / เริ่มแล้ว / จองซ้ำ / เต็ม ถูกปฏิเสธพร้อมเหตุผลและไม่บันทึก (ASM-03, ASM-05).
func TestStoreBookRejections(t *testing.T) {
	pool := newTestPool(t)
	store := NewStore(pool)
	ctx := context.Background()
	a, b := newMember(t, pool), newMember(t, pool)

	t.Run("ไม่พบคลาส", func(t *testing.T) {
		_, err := store.Book(ctx, 999_999, a, testNow, allowShift)
		assert.ErrorIs(t, err, ErrClassNotFound)
	})

	t.Run("ถึงเวลาเริ่มแล้ว (Q-12)", func(t *testing.T) {
		classID := newClass(t, pool, 1, span("08:00", "09:00"), 10)
		_, err := store.Book(ctx, classID, a, testNow, allowShift)
		assert.ErrorIs(t, err, ErrClassStarted)
		assert.Zero(t, countBookings(t, pool, classID))
	})

	t.Run("จองคลาสเดิมซ้ำ", func(t *testing.T) {
		classID := newClass(t, pool, 2, span("10:00", "11:00"), 10)
		_, err := store.Book(ctx, classID, a, testNow, allowShift)
		require.NoError(t, err)
		_, err = store.Book(ctx, classID, a, testNow, allowShift)
		assert.ErrorIs(t, err, ErrAlreadyBooked)
		assert.Equal(t, 1, countBookings(t, pool, classID))
	})

	t.Run("คลาสเต็ม", func(t *testing.T) {
		classID := newClass(t, pool, 3, span("10:00", "11:00"), 1)
		_, err := store.Book(ctx, classID, a, testNow, allowShift)
		require.NoError(t, err)
		_, err = store.Book(ctx, classID, b, testNow, allowShift)
		assert.ErrorIs(t, err, ErrClassFull)
		assert.Equal(t, 1, countBookings(t, pool, classID))
	})
}

// TestStoreShiftCheck: ตัวตรวจกะได้ trainer และช่วงเวลาของคลาส ถ้าตอบนอกกะจะไม่บันทึก และไม่ถูกเรียกถ้าคลาสเริ่มแล้ว (DOM-BKG-01, plan D3).
func TestStoreShiftCheck(t *testing.T) {
	pool := newTestPool(t)
	store := NewStore(pool)
	ctx := context.Background()
	member := newMember(t, pool)
	classID := newClass(t, pool, 7, span("10:00", "11:00"), 10)

	var gotTrainer int64
	var gotInterval Interval
	offShift := func(_ context.Context, trainerID int64, iv Interval) error {
		gotTrainer, gotInterval = trainerID, iv
		return ErrTrainerOffShift
	}
	_, err := store.Book(ctx, classID, member, testNow, offShift)
	assert.ErrorIs(t, err, ErrTrainerOffShift)
	assert.Equal(t, int64(7), gotTrainer)
	assert.True(t, gotInterval.Start.Equal(at("10:00")) && gotInterval.End.Equal(at("11:00")))
	assert.Zero(t, countBookings(t, pool, classID))

	readFails := errors.New("roster unavailable")
	_, err = store.Book(ctx, classID, member, testNow, func(context.Context, int64, Interval) error { return readFails })
	assert.ErrorIs(t, err, readFails)
	assert.Zero(t, countBookings(t, pool, classID))

	started := newClass(t, pool, 7, span("08:00", "09:00"), 10)
	called := false
	_, err = store.Book(ctx, started, member, testNow, func(context.Context, int64, Interval) error {
		called = true
		return ErrTrainerOffShift
	})
	assert.ErrorIs(t, err, ErrClassStarted, "เริ่มแล้วต้องมาก่อนนอกกะ (plan D3)")
	assert.False(t, called)
}

// TestStoreOverlapCountsOnlyBookedClasses: คลาสนับเป็นคิวสอนเมื่อมีผู้จองแล้ว (ASM-10) เฉพาะเทรนเนอร์คนเดียวกัน (ASM-09)
// และต่อกันพอดีนับว่าทับ (ASM-12).
func TestStoreOverlapCountsOnlyBookedClasses(t *testing.T) {
	pool := newTestPool(t)
	store := NewStore(pool)
	ctx := context.Background()
	a, b := newMember(t, pool), newMember(t, pool)

	x := newClass(t, pool, 1, span("10:00", "11:00"), 10)
	y := newClass(t, pool, 1, span("10:30", "11:30"), 10)
	next := newClass(t, pool, 1, span("11:30", "12:30"), 10)
	otherTrainer := newClass(t, pool, 2, span("10:00", "11:00"), 10)

	// X ยังไม่มีผู้จอง จึงจอง Y ที่ทับกับ X ได้
	_, err := store.Book(ctx, y, a, testNow, allowShift)
	require.NoError(t, err)

	// ตอนนี้ Y เป็นคิวสอนแล้ว จอง X ที่ทับกับ Y ไม่ได้
	_, err = store.Book(ctx, x, b, testNow, allowShift)
	assert.ErrorIs(t, err, ErrTrainerBusy)
	assert.Zero(t, countBookings(t, pool, x))

	// คลาสที่เริ่มตอน Y จบพอดี นับว่าทับ (ASM-12)
	_, err = store.Book(ctx, next, b, testNow, allowShift)
	assert.ErrorIs(t, err, ErrTrainerBusy)

	// เทรนเนอร์คนอื่นเวลาเดียวกันไม่เกี่ยว (ASM-09)
	_, err = store.Book(ctx, otherTrainer, b, testNow, allowShift)
	assert.NoError(t, err)

	// จองคลาส Y เดิมเพิ่มอีกคน ไม่นับว่าทับกับตัวเอง (ASM-10)
	_, err = store.Book(ctx, y, b, testNow, allowShift)
	assert.NoError(t, err)
}
