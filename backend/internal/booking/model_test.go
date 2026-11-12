package booking

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bangkok คือเขตเวลาของข้อมูลทดสอบ (เวลาใน AC เป็นเวลาท้องถิ่นของยิม).
var bangkok = time.FixedZone("Asia/Bangkok", 7*60*60)

// at คืนเวลา hh:mm ของวันทดสอบ 13 พฤศจิกายน 2026 (ตัวอย่างใน AC เป็นข้อมูลทดสอบ ไม่ใช่ข้อกำหนด).
func at(hhmm string) time.Time {
	var h, m int
	if _, err := fmt.Sscanf(hhmm, "%d:%d", &h, &m); err != nil {
		panic(err)
	}
	return time.Date(2026, 11, 13, h, m, 0, 0, bangkok)
}

// span คืน Interval จาก "hh:mm" ถึง "hh:mm".
func span(start, end string) Interval { return Interval{Start: at(start), End: at(end)} }

// TestAC_BKG_02_Rules: Given เทรนเนอร์มีคลาส X 10:00–11:00 When เทียบกับคลาส Y Then ทับซ้อนตามช่วงปิด (ASM-12).
func TestAC_BKG_02_Rules(t *testing.T) {
	x := span("10:00", "11:00")
	cases := []struct {
		name string
		y    Interval
		want bool
	}{
		{"Y 10:30–11:30 ทับ (ตัวอย่างใน AC-BKG-02)", span("10:30", "11:30"), true},
		{"Y 11:00–12:00 ต่อกันพอดี นับว่าทับ (ASM-12)", span("11:00", "12:00"), true},
		{"Y 09:00–10:00 ต่อกันพอดีด้านหน้า นับว่าทับ (ASM-12)", span("09:00", "10:00"), true},
		{"Y 10:15–10:45 อยู่ข้างใน ทับ", span("10:15", "10:45"), true},
		{"Y 11:01–12:00 ไม่ทับ", span("11:01", "12:00"), false},
		{"Y 08:00–09:59 ไม่ทับ", span("08:00", "09:59"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, x.Overlaps(tc.y))
			assert.Equal(t, tc.want, tc.y.Overlaps(x), "ต้องสมมาตร")
		})
	}
}

// TestAC_BKG_03_Rules: Given เทรนเนอร์มีกะ 09:00–12:00 When คลาสเริ่มก่อน 09:00 Then ไม่อยู่ในกะ; เริ่ม 09:00 พอดีอยู่ในกะ (ASM-13).
func TestAC_BKG_03_Rules(t *testing.T) {
	shift := span("09:00", "12:00")
	assert.False(t, span("08:30", "09:30").Within(shift), "เริ่มก่อน 09:00 ต้องนอกกะ (AC-BKG-03)")
	assert.False(t, span("08:59", "10:00").Within(shift), "เริ่มก่อน 09:00 แม้นาทีเดียวต้องนอกกะ")
	assert.True(t, span("09:00", "10:00").Within(shift), "เริ่มตรงต้นกะนับว่าอยู่ในกะ (ASM-13)")
}

// TestAC_BKG_04_Rules: Given เทรนเนอร์มีกะ 09:00–12:00 When คลาส 11:30–12:30 Then ไม่อยู่ในกะ; จบ 12:00 พอดีอยู่ในกะ (ASM-13).
func TestAC_BKG_04_Rules(t *testing.T) {
	shift := span("09:00", "12:00")
	assert.False(t, span("11:30", "12:30").Within(shift), "จบหลัง 12:00 ต้องนอกกะ (AC-BKG-04)")
	assert.True(t, span("11:00", "12:00").Within(shift), "จบตรงสิ้นสุดกะนับว่าอยู่ในกะ (ASM-13)")
	assert.True(t, span("09:00", "12:00").Within(shift), "เต็มกะพอดีนับว่าอยู่ในกะ")
}

// TestReasonCodes ยืนยันว่า error การจองทุกตัวมีรหัสเหตุผลตาม plan ข้อ 4 รวมเมื่อถูกห่อด้วย %w (ASM-05).
func TestReasonCodes(t *testing.T) {
	want := map[error]string{
		ErrClassNotFound:   "class_not_found",
		ErrClassStarted:    "class_started",
		ErrAlreadyBooked:   "already_booked",
		ErrClassFull:       "class_full",
		ErrTrainerOffShift: "trainer_off_shift",
		ErrTrainerBusy:     "trainer_busy",
		ErrBookingNotFound: "booking_not_found",
	}
	for err, code := range want {
		got, ok := ReasonCode(fmt.Errorf("wrapped: %w", err))
		require.True(t, ok, err.Error())
		assert.Equal(t, code, got)
	}
	_, ok := ReasonCode(fmt.Errorf("db down"))
	assert.False(t, ok)
}
