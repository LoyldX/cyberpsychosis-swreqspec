package roster

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tod แปลง "HH:MM" เป็น TimeOfDay สำหรับ test.
func tod(s string) TimeOfDay {
	t, err := ParseTimeOfDay(s)
	if err != nil {
		panic(err)
	}
	return t
}

// slot คือ Slot ของวัน weekday จาก "HH:MM" ถึง "HH:MM".
func slot(weekday int, start, end string) Slot {
	return Slot{Weekday: weekday, Start: tod(start), End: tod(end)}
}

// TestParseTimeOfDay: รับ "HH:MM" 00:00–23:59 และแสดงกลับเป็นรูปเดิม (ASM-03).
func TestParseTimeOfDay(t *testing.T) {
	for _, s := range []string{"00:00", "09:00", "12:30", "23:59"} {
		got, err := ParseTimeOfDay(s)
		require.NoError(t, err, s)
		assert.Equal(t, s, got.String())
	}
	for _, s := range []string{"", "9:00", "24:00", "12:60", "12:00:00", "noon"} {
		_, err := ParseTimeOfDay(s)
		assert.ErrorIs(t, err, ErrInvalidEntry, s)
	}
}

// TestParseKind: ประเภทรายการมีแค่เวลาเข้างานและเวลาสอน (ASM-04).
func TestParseKind(t *testing.T) {
	k, err := ParseKind("work")
	require.NoError(t, err)
	assert.Equal(t, KindWork, k)
	k, err = ParseKind("teach")
	require.NoError(t, err)
	assert.Equal(t, KindTeach, k)
	_, err = ParseKind("lunch")
	assert.ErrorIs(t, err, ErrInvalidEntry)
}

// TestSlotValidate: วัน 1..7 และเวลาเริ่มต้องก่อนเวลาจบในวันเดียว (ASM-02, ASM-03, Q-07).
func TestSlotValidate(t *testing.T) {
	assert.NoError(t, slot(1, "09:00", "12:00").Validate())
	assert.NoError(t, slot(7, "00:00", "23:59").Validate())
	assert.ErrorIs(t, slot(0, "09:00", "12:00").Validate(), ErrInvalidEntry)
	assert.ErrorIs(t, slot(8, "09:00", "12:00").Validate(), ErrInvalidEntry)
	assert.ErrorIs(t, slot(1, "12:00", "12:00").Validate(), ErrInvalidEntry, "เริ่มเท่ากับจบ")
	assert.ErrorIs(t, slot(5, "20:00", "02:00").Validate(), ErrInvalidEntry, "ข้ามเที่ยงคืน")
}

// TestSlotOverlaps: ทับซ้อนเมื่อวันเดียวกันและคาบเกี่ยวจริง ต่อกันพอดีไม่นับ (ASM-02, ASM-10).
func TestSlotOverlaps(t *testing.T) {
	a := slot(1, "09:00", "12:00")
	assert.True(t, a.Overlaps(slot(1, "11:00", "13:00")))
	assert.True(t, a.Overlaps(slot(1, "10:00", "11:00")))
	assert.False(t, a.Overlaps(slot(1, "12:00", "15:00")), "ต่อกันพอดีด้านหลังไม่นับ (ASM-10)")
	assert.False(t, a.Overlaps(slot(1, "07:00", "09:00")), "ต่อกันพอดีด้านหน้าไม่นับ (ASM-10)")
	assert.False(t, a.Overlaps(slot(2, "10:00", "11:00")), "คนละวัน")
}

// TestSlotWithin: อยู่ในเวลาเข้างานเมื่ออยู่ในช่วงเดียวหรือช่วงที่ต่อกัน ขอบพอดีนับว่าอยู่ใน (ASM-04, ASM-10).
func TestSlotWithin(t *testing.T) {
	work := []Slot{slot(1, "09:00", "12:00"), slot(1, "12:00", "15:00"), slot(1, "16:00", "18:00"), slot(2, "09:00", "12:00")}
	assert.True(t, slot(1, "09:00", "12:00").Within(work), "ขอบพอดี")
	assert.True(t, slot(1, "11:30", "12:30").Within(work), "คร่อมสองช่วงที่ต่อกัน (ASM-10)")
	assert.True(t, slot(1, "09:00", "15:00").Within(work))
	assert.False(t, slot(1, "14:30", "16:30").Within(work), "คร่อมช่องว่าง 15:00–16:00")
	assert.False(t, slot(1, "08:30", "09:30").Within(work), "เริ่มก่อนเวลาเข้างาน")
	assert.False(t, slot(3, "09:00", "10:00").Within(work), "วันที่ไม่มีเวลาเข้างาน")
	assert.False(t, slot(1, "09:00", "10:00").Within(nil))
}

// TestSlotOf: แปลงเวลาจริงเป็นวัน+เวลา Asia/Bangkok และช่วงที่ข้ามวันไม่ได้เป็น Slot (ASM-03; 005 ASM-11).
func TestSlotOf(t *testing.T) {
	// 13 พ.ย. 2026 เป็นวันศุกร์ (ISO 5)
	start := time.Date(2026, 11, 13, 10, 0, 0, 0, bangkok)
	got, ok := SlotOf(start, start.Add(time.Hour))
	require.True(t, ok)
	assert.Equal(t, slot(5, "10:00", "11:00"), got)

	// เวลา UTC ถูกแปลงเป็นเวลาไทยก่อน: 03:00Z = 10:00 ไทย
	got, ok = SlotOf(start.UTC(), start.Add(time.Hour).UTC())
	require.True(t, ok)
	assert.Equal(t, slot(5, "10:00", "11:00"), got)

	// วันอาทิตย์เป็น 7
	sunday := time.Date(2026, 11, 15, 9, 0, 0, 0, bangkok)
	got, ok = SlotOf(sunday, sunday.Add(30*time.Minute))
	require.True(t, ok)
	assert.Equal(t, 7, got.Weekday)

	// ข้ามเที่ยงคืน
	late := time.Date(2026, 11, 13, 23, 30, 0, 0, bangkok)
	_, ok = SlotOf(late, late.Add(time.Hour))
	assert.False(t, ok)
}

// TestReasonCodes: error ทุกตัวมีรหัสเหตุผลตาม plan ข้อ 4 รวมเมื่อถูกห่อ (ASM-02).
func TestReasonCodes(t *testing.T) {
	want := map[error]string{
		ErrInvalidEntry:           "invalid_request",
		ErrTrainerNotFound:        "trainer_not_found",
		ErrEntryNotFound:          "entry_not_found",
		ErrOverlap:                "overlap",
		ErrTeachOutsideWork:       "teach_outside_work",
		ErrBookedClassOutsideWork: "booked_class_outside_work",
	}
	for err, code := range want {
		got, ok := ReasonCode(fmt.Errorf("wrapped: %w", err))
		require.True(t, ok, err.Error())
		assert.Equal(t, code, got)
	}
	_, ok := ReasonCode(fmt.Errorf("db down"))
	assert.False(t, ok)
}
