package booking

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixedClock คือ Clock จำลองที่คืนเวลาเดียวเสมอ.
type fixedClock struct{ now time.Time }

// Now คืนเวลาที่กำหนดไว้.
func (c fixedClock) Now() time.Time { return c.now }

// fakeShifts คือ ShiftChecker จำลอง: กะของเทรนเนอร์แต่ละคน (เทรนเนอร์ที่ไม่มีในแผนที่ = ไม่มีข้อมูลกะ)
// และ err ใช้จำลองการอ่านข้อมูลกะล้มเหลว (ASM-11, ASM-13).
type fakeShifts struct {
	byTrainer map[int64]Interval
	err       error
}

// Covers ใช้ Interval.Within กับกะของเทรนเนอร์ (ขอบพอดีนับว่าอยู่ในกะ ASM-13).
func (f fakeShifts) Covers(_ context.Context, trainerID int64, class Interval) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	shift, ok := f.byTrainer[trainerID]
	if !ok {
		return false, nil
	}
	return class.Within(shift), nil
}

// morningShift คือกะ 09:00–12:00 ของเทรนเนอร์ 1 ตามตัวอย่างใน AC-BKG-03/04.
var morningShift = fakeShifts{byTrainer: map[int64]Interval{1: span("09:00", "12:00")}}

// TestServiceRejectsOffShift: คลาสนอกกะถูกปฏิเสธด้วย ErrTrainerOffShift และไม่บันทึก ส่วนในกะจองได้ (DOM-BKG-01, ASM-08).
func TestServiceRejectsOffShift(t *testing.T) {
	pool := newTestPool(t)
	member := newMember(t, pool)
	svc := NewService(NewStore(pool), morningShift, fixedClock{testNow})
	ctx := context.Background()

	early := newClass(t, pool, 1, span("08:30", "09:30"), 10)
	_, err := svc.Book(ctx, early, member)
	assert.ErrorIs(t, err, ErrTrainerOffShift)
	assert.Zero(t, countBookings(t, pool, early))

	inShift := newClass(t, pool, 1, span("11:00", "12:00"), 10)
	b, err := svc.Book(ctx, inShift, member)
	require.NoError(t, err)
	assert.Equal(t, inShift, b.ClassID)
	assert.Equal(t, member, b.MemberID)
}

// TestServiceWithoutShiftDataBlocks: เทรนเนอร์ที่ไม่มีข้อมูลกะ หรือ Service ที่ไม่มีแหล่งข้อมูลกะ ถือว่านอกกะ (plan D6).
func TestServiceWithoutShiftDataBlocks(t *testing.T) {
	pool := newTestPool(t)
	member := newMember(t, pool)
	ctx := context.Background()
	classID := newClass(t, pool, 2, span("10:00", "11:00"), 10) // เทรนเนอร์ 2 ไม่มีกะใน morningShift

	_, err := NewService(NewStore(pool), morningShift, fixedClock{testNow}).Book(ctx, classID, member)
	assert.ErrorIs(t, err, ErrTrainerOffShift)

	_, err = NewService(NewStore(pool), nil, fixedClock{testNow}).Book(ctx, classID, member)
	assert.ErrorIs(t, err, ErrTrainerOffShift)
	assert.Zero(t, countBookings(t, pool, classID))
}

// TestServiceShiftReadFailure: อ่านข้อมูลกะไม่ได้ ไม่บันทึก และ error ไม่ใช่เหตุผลการจอง (handler จะตอบ 503 ตาม ASM-07).
func TestServiceShiftReadFailure(t *testing.T) {
	pool := newTestPool(t)
	member := newMember(t, pool)
	classID := newClass(t, pool, 1, span("10:00", "11:00"), 10)
	readFails := errors.New("roster unavailable")

	_, err := NewService(NewStore(pool), fakeShifts{err: readFails}, fixedClock{testNow}).Book(context.Background(), classID, member)
	assert.ErrorIs(t, err, readFails)
	_, isReason := ReasonCode(err)
	assert.False(t, isReason)
	assert.Zero(t, countBookings(t, pool, classID))
}

// TestServiceUsesClock: "ตอนนี้" มาจาก Clock ทั้งรายการคลาสและการตรวจคลาสเริ่มแล้ว (ASM-05, plan D4).
func TestServiceUsesClock(t *testing.T) {
	pool := newTestPool(t)
	member := newMember(t, pool)
	ctx := context.Background()
	classID := newClass(t, pool, 1, span("10:00", "11:00"), 10)

	before := NewService(NewStore(pool), morningShift, fixedClock{at("09:59")})
	classes, err := before.Classes(ctx)
	require.NoError(t, err)
	require.Len(t, classes, 1)

	atStart := NewService(NewStore(pool), morningShift, fixedClock{at("10:00")})
	classes, err = atStart.Classes(ctx)
	require.NoError(t, err)
	assert.Empty(t, classes)
	_, err = atStart.Book(ctx, classID, member)
	assert.ErrorIs(t, err, ErrClassStarted)
}
