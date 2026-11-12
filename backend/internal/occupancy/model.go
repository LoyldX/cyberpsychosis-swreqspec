// Package occupancy นับจำนวนผู้ใช้บริการภายในยิมจากสัญญาณผ่านประตู และส่งค่าให้ Dashboard (FR-BI-01, NFR-PERF-02, DOM-CAP-01, IF-GATE-01).
package occupancy

import (
	"errors"
	"time"
	_ "time/tzdata" // ฝังข้อมูล timezone ให้ Asia/Bangkok ใช้ได้แม้ระบบไม่มี tzdata (ASM-03, plan D7)
)

// Direction คือทิศทางของสัญญาณผ่านประตู 1 ครั้ง (IF-GATE-01, ASM-01: ประตูบันทึกทั้งขาเข้าและขาออก).
type Direction string

const (
	// DirectionIn คือผ่านเข้า (+1 ตาม DOM-CAP-01).
	DirectionIn Direction = "in"
	// DirectionOut คือผ่านออก (-1 ตาม DOM-CAP-01 แต่ไม่ต่ำกว่า 0 ตาม ASM-04).
	DirectionOut Direction = "out"
)

// ErrInvalidDirection คือสัญญาณที่ทิศทางไม่ใช่ "in" หรือ "out" (IF-GATE-01).
var ErrInvalidDirection = errors.New("occupancy: invalid direction")

// ParseDirection แปลงข้อความเป็น Direction และปฏิเสธค่าอื่นนอกจาก "in"/"out" (IF-GATE-01).
func ParseDirection(s string) (Direction, error) {
	switch Direction(s) {
	case DirectionIn, DirectionOut:
		return Direction(s), nil
	default:
		return "", ErrInvalidDirection
	}
}

// ApplySignal คืนจำนวนปัจจุบันหลังรับสัญญาณ 1 ครั้ง: เข้า +1, ออก -1 แต่ไม่ต่ำกว่า 0 (DOM-CAP-01, ASM-04).
func ApplySignal(current int, d Direction) int {
	if d == DirectionIn {
		return current + 1
	}
	if current <= 0 {
		return 0
	}
	return current - 1
}

// bangkok คือเขตเวลาที่ใช้แบ่ง "วันปัจจุบัน" (ASM-03).
var bangkok = mustLoadLocation("Asia/Bangkok")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic("occupancy: load location " + name + ": " + err.Error())
	}
	return loc
}

// StartOfDay คืนเวลา 00:00 Asia/Bangkok ของวันที่ t อยู่ — จำนวนปัจจุบันนับเฉพาะสัญญาณตั้งแต่เวลานี้ (ASM-03, DOM-CAP-01).
func StartOfDay(t time.Time) time.Time {
	local := t.In(bangkok)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, bangkok)
}

// NextStartOfDay คืนเวลา 00:00 Asia/Bangkok ของวันถัดไปหลัง t — เวลาที่จำนวนปัจจุบันเริ่มนับใหม่ (ASM-03).
func NextStartOfDay(t time.Time) time.Time {
	return StartOfDay(t).AddDate(0, 0, 1)
}
