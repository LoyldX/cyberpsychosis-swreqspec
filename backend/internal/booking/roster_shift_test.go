package booking

import (
	"context"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gymapp/backend/internal/roster"
)

// newRosterTrainer สร้างเทรนเนอร์ใน trainers ของ 004 พร้อมเวลาเข้างานศุกร์ (ISO 5) 09:00–12:00 ผ่าน roster.Store
// แล้วลบรายการและเทรนเนอร์นั้นตอนจบ test (ASM-11; วันทดสอบ 13 พ.ย. 2026 เป็นวันศุกร์).
func newRosterTrainer(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()
	ctx := context.Background()
	var trainerID int64
	require.NoError(t, pool.QueryRow(ctx, `INSERT INTO trainers (name) VALUES ('เทรนเนอร์ 005') RETURNING id`).Scan(&trainerID))
	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), `DELETE FROM roster_entries WHERE trainer_id = $1`, trainerID)
		require.NoError(t, err)
		_, err = pool.Exec(context.Background(), `DELETE FROM trainers WHERE id = $1`, trainerID)
		require.NoError(t, err)
	})
	start, err := roster.ParseTimeOfDay("09:00")
	require.NoError(t, err)
	end, err := roster.ParseTimeOfDay("12:00")
	require.NoError(t, err)
	_, err = roster.NewStore(pool).Create(ctx, trainerID, roster.KindWork, roster.Slot{Weekday: 5, Start: start, End: end})
	require.NoError(t, err)
	return trainerID
}

// TestAC_BKG_03_Roster: Given เทรนเนอร์มีกะ 09:00–12:00 (เวลาเข้างานจริงในตารางงานของ 004) When สมาชิกจองคลาสที่เริ่มก่อน 09:00
// Then ระบบบล็อกการจอง; คลาสที่เริ่ม 09:00 พอดีจองได้ (FR-BKG-02, DOM-BKG-01; ASM-08, ASM-11, ASM-13).
func TestAC_BKG_03_Roster(t *testing.T) {
	pool := newTestPool(t)
	member := newMember(t, pool)
	trainer := newRosterTrainer(t, pool)
	early := newClass(t, pool, trainer, span("08:30", "09:30"), 10)
	atStart := newClass(t, pool, trainer, span("09:00", "10:00"), 10)
	r := newDBRouter(pool, NewRosterShifts(roster.NewStore(pool)), member)

	requireReason(t, postBooking(r, member, early), http.StatusConflict, CodeTrainerOffShift)
	assert.Zero(t, countBookings(t, pool, early))
	require.Equal(t, http.StatusCreated, postBooking(r, member, atStart).Code)
}

// TestAC_BKG_04_Roster: Given เทรนเนอร์มีกะ 09:00–12:00 (เวลาเข้างานจริงในตารางงานของ 004) When สมาชิกจองคลาสที่เริ่ม 11:30 และจบ 12:30
// Then ระบบบล็อกการจอง; คลาสที่จบ 12:00 พอดีจองได้ (FR-BKG-02, DOM-BKG-01; ASM-08, ASM-11, ASM-13).
func TestAC_BKG_04_Roster(t *testing.T) {
	pool := newTestPool(t)
	member := newMember(t, pool)
	trainer := newRosterTrainer(t, pool)
	late := newClass(t, pool, trainer, span("11:30", "12:30"), 10)
	atEnd := newClass(t, pool, trainer, span("11:00", "12:00"), 10)
	r := newDBRouter(pool, NewRosterShifts(roster.NewStore(pool)), member)

	requireReason(t, postBooking(r, member, late), http.StatusConflict, CodeTrainerOffShift)
	assert.Zero(t, countBookings(t, pool, late))
	require.Equal(t, http.StatusCreated, postBooking(r, member, atEnd).Code)
}
