// Package booking คือการจองคลาสออกกำลังกายของสมาชิก (FR-BKG-01, FR-BKG-02, NFR-PERF-01, DOM-BKG-01, CON-TECH-01).
package booking

import (
	"errors"
	"time"
)

// Interval คือช่วงเวลา [Start, End] ของคลาสหรือกะ ใช้เทียบแบบช่วงปิด (DOM-BKG-01, ASM-12, ASM-13).
type Interval struct {
	Start time.Time
	End   time.Time
}

// Overlaps บอกว่าสองช่วงเวลาทับซ้อนกันหรือไม่ แบบช่วงปิด: ต่อกันพอดี (จบ 11:00 อีกช่วงเริ่ม 11:00) นับว่าทับ
// (FR-BKG-02, DOM-BKG-01, ASM-12).
func (a Interval) Overlaps(b Interval) bool {
	return !a.Start.After(b.End) && !b.Start.After(a.End)
}

// Within บอกว่าช่วงเวลา a อยู่ภายใน outer ทั้งหมดหรือไม่: เริ่มไม่ก่อนต้น outer และจบไม่เกินท้าย outer
// ขอบพอดีนับว่าอยู่ภายใน (FR-BKG-02, DOM-BKG-01, ASM-13).
func (a Interval) Within(outer Interval) bool {
	return !a.Start.Before(outer.Start) && !a.End.After(outer.End)
}

// รหัสเหตุผลที่ API ส่งกลับในช่อง "error" ตาม plan ข้อ 4 (FR-BKG-01, FR-BKG-02, ASM-05, ASM-07).
const (
	CodeInvalidRequest  = "invalid_request"
	CodeClassNotFound   = "class_not_found"
	CodeClassStarted    = "class_started"
	CodeAlreadyBooked   = "already_booked"
	CodeClassFull       = "class_full"
	CodeTrainerOffShift = "trainer_off_shift"
	CodeTrainerBusy     = "trainer_busy"
	CodeBookingNotFound = "booking_not_found"
	CodeUnavailable     = "unavailable"
)

// error ของการจองที่ไม่ผ่านเงื่อนไข แต่ละตัวผูกกับรหัสเหตุผลด้วย ReasonCode (ASM-05, ASM-08).
var (
	ErrClassNotFound   = errors.New("booking: class not found")
	ErrClassStarted    = errors.New("booking: class already started")
	ErrAlreadyBooked   = errors.New("booking: member already booked this class")
	ErrClassFull       = errors.New("booking: class is full")
	ErrTrainerOffShift = errors.New("booking: class is outside trainer shift")
	ErrTrainerBusy     = errors.New("booking: trainer has another booked class at this time")
	ErrBookingNotFound = errors.New("booking: booking not found")
)

// reasonCodes จับคู่ error กับรหัสเหตุผลของ API (ASM-05).
var reasonCodes = map[error]string{
	ErrClassNotFound:   CodeClassNotFound,
	ErrClassStarted:    CodeClassStarted,
	ErrAlreadyBooked:   CodeAlreadyBooked,
	ErrClassFull:       CodeClassFull,
	ErrTrainerOffShift: CodeTrainerOffShift,
	ErrTrainerBusy:     CodeTrainerBusy,
	ErrBookingNotFound: CodeBookingNotFound,
}

// ReasonCode คืนรหัสเหตุผลของ error การจอง และ false ถ้าไม่ใช่ error ที่รู้จัก (ASM-05).
func ReasonCode(err error) (string, bool) {
	for target, code := range reasonCodes {
		if errors.Is(err, target) {
			return code, true
		}
	}
	return "", false
}
