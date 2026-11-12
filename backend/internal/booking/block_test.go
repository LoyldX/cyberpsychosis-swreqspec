package booking

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAC_BKG_02_HTTP: Given เทรนเนอร์มีคลาส X 10:00–11:00 ที่มีสมาชิกจองแล้วอย่างน้อย 1 คน (ASM-10)
// When สมาชิกจองคลาส Y ของเทรนเนอร์เดียวกันเวลา 10:30–11:30 Then ระบบบล็อกการจองและไม่บันทึกการจอง Y
// (FR-BKG-02, DOM-BKG-01; ASM-09, ASM-12).
func TestAC_BKG_02_HTTP(t *testing.T) {
	shifts := fakeShifts{byTrainer: map[int64]Interval{1: span("08:30", "13:00")}} // อยู่ในกะทุกคลาส ตรวจเฉพาะทับซ้อน

	t.Run("ทับกับคลาสที่มีผู้จองแล้วถูกบล็อก", func(t *testing.T) {
		pool := newTestPool(t)
		first, second := newMember(t, pool), newMember(t, pool)
		x := newClass(t, pool, 1, span("10:00", "11:00"), 10)
		y := newClass(t, pool, 1, span("10:30", "11:30"), 10)
		r := newDBRouter(pool, shifts, first, second)

		require.Equal(t, http.StatusCreated, postBooking(r, first, x).Code) // Given: X มีผู้จองแล้ว
		requireReason(t, postBooking(r, second, y), http.StatusConflict, CodeTrainerBusy)
		assert.Zero(t, countBookings(t, pool, y))
	})

	t.Run("คลาสที่ต่อกันพอดีนับว่าทับซ้อน (ASM-12)", func(t *testing.T) {
		pool := newTestPool(t)
		first, second := newMember(t, pool), newMember(t, pool)
		x := newClass(t, pool, 1, span("10:00", "11:00"), 10)
		next := newClass(t, pool, 1, span("11:00", "12:00"), 10)
		r := newDBRouter(pool, shifts, first, second)

		require.Equal(t, http.StatusCreated, postBooking(r, first, x).Code)
		requireReason(t, postBooking(r, second, next), http.StatusConflict, CodeTrainerBusy)
		assert.Zero(t, countBookings(t, pool, next))
	})

	t.Run("ทับกับคลาสที่ยังไม่มีผู้จองไม่ถูกบล็อก (ASM-10)", func(t *testing.T) {
		pool := newTestPool(t)
		member := newMember(t, pool)
		newClass(t, pool, 1, span("10:00", "11:00"), 10) // X ยังไม่มีผู้จอง
		y := newClass(t, pool, 1, span("10:30", "11:30"), 10)
		r := newDBRouter(pool, shifts, member)

		require.Equal(t, http.StatusCreated, postBooking(r, member, y).Code)
		assert.Equal(t, 1, countBookings(t, pool, y))
	})
}

// TestAC_BKG_03_HTTP: Given เทรนเนอร์มีกะ 09:00–12:00 When สมาชิกจองคลาสที่เริ่มก่อน 09:00 Then ระบบบล็อกการจอง
// และไม่บันทึก; คลาสที่เริ่ม 09:00 พอดีจองได้ (FR-BKG-02, DOM-BKG-01; ASM-08, ASM-13).
func TestAC_BKG_03_HTTP(t *testing.T) {
	pool := newTestPool(t)
	member := newMember(t, pool)
	early := newClass(t, pool, 1, span("08:30", "09:30"), 10)
	atShiftStart := newClass(t, pool, 1, span("09:00", "10:00"), 10)
	r := newDBRouter(pool, morningShift, member)

	requireReason(t, postBooking(r, member, early), http.StatusConflict, CodeTrainerOffShift)
	assert.Zero(t, countBookings(t, pool, early))

	require.Equal(t, http.StatusCreated, postBooking(r, member, atShiftStart).Code)
}

// TestAC_BKG_04_HTTP: Given เทรนเนอร์มีกะ 09:00–12:00 When สมาชิกจองคลาสที่เริ่ม 11:30 และจบ 12:30 Then ระบบบล็อกการจอง
// และไม่บันทึก; คลาสที่จบ 12:00 พอดีจองได้ (FR-BKG-02, DOM-BKG-01; ASM-08, ASM-13).
func TestAC_BKG_04_HTTP(t *testing.T) {
	pool := newTestPool(t)
	member := newMember(t, pool)
	late := newClass(t, pool, 1, span("11:30", "12:30"), 10)
	atShiftEnd := newClass(t, pool, 1, span("11:00", "12:00"), 10)
	r := newDBRouter(pool, morningShift, member)

	requireReason(t, postBooking(r, member, late), http.StatusConflict, CodeTrainerOffShift)
	assert.Zero(t, countBookings(t, pool, late))

	require.Equal(t, http.StatusCreated, postBooking(r, member, atShiftEnd).Code)
}
