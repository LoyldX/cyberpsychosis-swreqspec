package occupancy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestAC_BI_02_ApplySignal: Given จำนวนปัจจุบัน 10 When รับสัญญาณผ่านเข้า 1 ครั้ง Then จำนวนเป็น 11 (DOM-CAP-01).
func TestAC_BI_02_ApplySignal(t *testing.T) {
	current := 10
	got := ApplySignal(current, DirectionIn)
	require.Equal(t, 11, got)
}

// TestAC_BI_03_ApplySignal: Given จำนวนปัจจุบัน 10 When รับสัญญาณผ่านออก 1 ครั้ง Then จำนวนเป็น 9 (DOM-CAP-01, ASM-01).
func TestAC_BI_03_ApplySignal(t *testing.T) {
	current := 10
	got := ApplySignal(current, DirectionOut)
	require.Equal(t, 9, got)
}

// TestAC_BI_05_ApplySignal: Given จำนวนปัจจุบัน 0 When รับสัญญาณผ่านออก 1 ครั้ง Then จำนวนยังเป็น 0 (DOM-CAP-01, ASM-04).
func TestAC_BI_05_ApplySignal(t *testing.T) {
	current := 0
	got := ApplySignal(current, DirectionOut)
	require.Equal(t, 0, got)
}

// TestParseDirection ยืนยันว่ารับเฉพาะ "in"/"out" (IF-GATE-01).
func TestParseDirection(t *testing.T) {
	d, err := ParseDirection("in")
	require.NoError(t, err)
	require.Equal(t, DirectionIn, d)
	d, err = ParseDirection("out")
	require.NoError(t, err)
	require.Equal(t, DirectionOut, d)
	for _, bad := range []string{"", "IN", "sideways"} {
		_, err = ParseDirection(bad)
		require.ErrorIs(t, err, ErrInvalidDirection, bad)
	}
}

// TestStartOfDayBangkok ยืนยันว่า 23:59 และ 00:00 Asia/Bangkok อยู่คนละวัน แม้เป็นวันเดียวกันใน UTC (ASM-03).
func TestStartOfDayBangkok(t *testing.T) {
	bkk, err := time.LoadLocation("Asia/Bangkok")
	require.NoError(t, err)
	beforeMidnight := time.Date(2026, 11, 12, 23, 59, 0, 0, bkk)
	atMidnight := time.Date(2026, 11, 13, 0, 0, 0, 0, bkk)

	require.Equal(t, time.Date(2026, 11, 12, 0, 0, 0, 0, bkk), StartOfDay(beforeMidnight))
	require.Equal(t, atMidnight, StartOfDay(atMidnight))
	require.NotEqual(t, StartOfDay(beforeMidnight), StartOfDay(atMidnight))
	require.Equal(t, atMidnight, NextStartOfDay(beforeMidnight))

	// 16:59 UTC = 23:59 ที่กรุงเทพ ต้องได้จุดเริ่มวันแบบกรุงเทพ ไม่ใช่ UTC
	require.True(t, StartOfDay(beforeMidnight.UTC()).Equal(time.Date(2026, 11, 12, 0, 0, 0, 0, bkk)))
}
