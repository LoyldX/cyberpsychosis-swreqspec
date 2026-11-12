// Package roster คือตารางงาน (เวลาเข้างานและเวลาสอน) ของเทรนเนอร์ (FR-ROST-01, CON-DB-01, CON-TECH-01).
package roster

import (
	"errors"
	"fmt"
	"sort"
	"time"
	_ "time/tzdata" // ให้หา Asia/Bangkok ได้แม้เครื่องไม่มีฐานข้อมูลเขตเวลา (ASM-03)
)

// Kind คือประเภทรายการ: เวลาเข้างานหรือเวลาสอน ซึ่งเป็นข้อมูลแยกกัน (FR-ROST-01, ASM-04).
type Kind string

const (
	KindWork  Kind = "work"  // เวลาเข้างาน
	KindTeach Kind = "teach" // เวลาสอน
)

// ParseKind ตรวจว่าเป็น work หรือ teach (ASM-04).
func ParseKind(s string) (Kind, error) {
	switch Kind(s) {
	case KindWork, KindTeach:
		return Kind(s), nil
	}
	return "", fmt.Errorf("%w: kind %q", ErrInvalidEntry, s)
}

// TimeOfDay คือเวลาในวันเป็นวินาทีนับจาก 00:00 (ตามเวลา Asia/Bangkok) (ASM-03).
type TimeOfDay int

const secondsPerDay = 24 * 60 * 60

// ParseTimeOfDay แปลง "HH:MM" เป็น TimeOfDay (00:00–23:59) (ASM-03).
func ParseTimeOfDay(s string) (TimeOfDay, error) {
	t, err := time.Parse("15:04", s)
	if err != nil || len(s) != 5 {
		return 0, fmt.Errorf("%w: time %q", ErrInvalidEntry, s)
	}
	return TimeOfDay(t.Hour()*3600 + t.Minute()*60), nil
}

// String แสดงเป็น "HH:MM" (หรือ "HH:MM:SS" ถ้ามีวินาที).
func (t TimeOfDay) String() string {
	h, m, s := int(t)/3600, int(t)%3600/60, int(t)%60
	if s != 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", h, m)
}

// Slot คือช่วงเวลาซ้ำทุกสัปดาห์: วัน (ISO จันทร์ = 1 … อาทิตย์ = 7) + เวลาเริ่ม-จบในวันเดียว (ASM-03).
type Slot struct {
	Weekday int
	Start   TimeOfDay
	End     TimeOfDay
}

// Validate ตรวจว่าวันอยู่ใน 1..7 และเวลาเริ่มก่อนเวลาจบในวันเดียว ไม่ข้ามเที่ยงคืน (ASM-02, ASM-03).
func (s Slot) Validate() error {
	if s.Weekday < 1 || s.Weekday > 7 {
		return fmt.Errorf("%w: weekday %d", ErrInvalidEntry, s.Weekday)
	}
	if s.Start < 0 || s.End > secondsPerDay || s.Start >= s.End {
		return fmt.Errorf("%w: start %s must be before end %s", ErrInvalidEntry, s.Start, s.End)
	}
	return nil
}

// Overlaps บอกว่าสองช่วงทับซ้อนกันหรือไม่ — ต้องเป็นวันเดียวกันและคาบเกี่ยวกันจริง ต่อกันพอดีไม่นับ (ASM-02, ASM-10).
func (s Slot) Overlaps(o Slot) bool {
	return s.Weekday == o.Weekday && s.Start < o.End && o.Start < s.End
}

// Within บอกว่า s อยู่ในเวลาเข้างาน work ทั้งหมดหรือไม่ โดยรวมช่วงที่ต่อกันหรือทับกันเป็นช่วงเดียว
// ขอบพอดีนับว่าอยู่ใน (ASM-04, ASM-10; 005 ASM-13).
func (s Slot) Within(work []Slot) bool {
	var day []Slot
	for _, w := range work {
		if w.Weekday == s.Weekday {
			day = append(day, w)
		}
	}
	sort.Slice(day, func(i, j int) bool { return day[i].Start < day[j].Start })
	for i := 0; i < len(day); {
		start, end := day[i].Start, day[i].End
		j := i + 1
		for j < len(day) && day[j].Start <= end { // ต่อกันพอดีหรือทับกันรวมเป็นช่วงเดียว (ASM-10)
			if day[j].End > end {
				end = day[j].End
			}
			j++
		}
		if start <= s.Start && s.End <= end {
			return true
		}
		i = j
	}
	return false
}

// bangkok คือเขตเวลาของตารางงาน (ASM-03).
var bangkok = mustLoadLocation("Asia/Bangkok")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

// SlotOf แปลงช่วงเวลาจริง (เช่นเวลาของคลาสใน 005) เป็น Slot ตามเวลา Asia/Bangkok
// ok เป็น false ถ้าช่วงเวลาข้ามวัน เพราะกะเริ่ม-จบในวันเดียว จึงไม่มีกะใดครอบได้ (ASM-03; 005 ASM-11).
func SlotOf(start, end time.Time) (Slot, bool) {
	s, e := start.In(bangkok), end.In(bangkok)
	sy, sm, sd := s.Date()
	ey, em, ed := e.Date()
	if sy != ey || sm != em || sd != ed || !s.Before(e) {
		return Slot{}, false
	}
	weekday := int(s.Weekday())
	if weekday == 0 { // time.Sunday = 0 → ISO 7
		weekday = 7
	}
	return Slot{Weekday: weekday, Start: secondsOf(s), End: secondsOf(e)}, true
}

func secondsOf(t time.Time) TimeOfDay {
	return TimeOfDay(t.Hour()*3600 + t.Minute()*60 + t.Second())
}

// รหัสเหตุผลที่ API ส่งกลับในช่อง "error" ตาม plan ข้อ 4 (FR-ROST-01, ASM-02, ASM-04, ASM-07, ASM-08).
const (
	CodeInvalidRequest         = "invalid_request"
	CodeTrainerNotFound        = "trainer_not_found"
	CodeEntryNotFound          = "entry_not_found"
	CodeOverlap                = "overlap"
	CodeTeachOutsideWork       = "teach_outside_work"
	CodeBookedClassOutsideWork = "booked_class_outside_work"
	CodeUnavailable            = "unavailable"
)

// error ของการเปลี่ยนแปลงตารางที่ไม่ผ่านเงื่อนไข (ASM-02, ASM-04, ASM-07).
var (
	ErrInvalidEntry           = errors.New("roster: invalid entry")
	ErrTrainerNotFound        = errors.New("roster: trainer not found")
	ErrEntryNotFound          = errors.New("roster: entry not found")
	ErrOverlap                = errors.New("roster: overlaps another entry")
	ErrTeachOutsideWork       = errors.New("roster: teaching time outside working time")
	ErrBookedClassOutsideWork = errors.New("roster: booked class would be outside working time")
)

// reasonCodes จับคู่ error กับรหัสเหตุผลของ API.
var reasonCodes = map[error]string{
	ErrInvalidEntry:           CodeInvalidRequest,
	ErrTrainerNotFound:        CodeTrainerNotFound,
	ErrEntryNotFound:          CodeEntryNotFound,
	ErrOverlap:                CodeOverlap,
	ErrTeachOutsideWork:       CodeTeachOutsideWork,
	ErrBookedClassOutsideWork: CodeBookedClassOutsideWork,
}

// ReasonCode คืนรหัสเหตุผลของ error และ false ถ้าไม่ใช่ error ที่รู้จัก (ASM-02).
func ReasonCode(err error) (string, bool) {
	for target, code := range reasonCodes {
		if errors.Is(err, target) {
			return code, true
		}
	}
	return "", false
}
