package booking

import (
	"context"
	"fmt"
	"time"
)

// Clock ให้เวลาปัจจุบัน แยกออกมาเพื่อให้ test กำหนดเวลาได้ (ASM-05 "ถึงเวลาเริ่มแล้ว", plan D4).
type Clock interface {
	Now() time.Time
}

// SystemClock คือ Clock จากนาฬิกาเครื่อง.
type SystemClock struct{}

// Now คืนเวลาปัจจุบันของเครื่อง.
func (SystemClock) Now() time.Time { return time.Now() }

// ShiftChecker ตอบว่าช่วงเวลาของคลาสอยู่ในเวลาเข้างานของเทรนเนอร์ทั้งหมดหรือไม่ (ขอบพอดีนับว่าอยู่ในกะ)
// ตัวจริงอ่านจากตารางงานของ 004 (T-10) (FR-BKG-02, DOM-BKG-01, ASM-11, ASM-13, plan D6).
type ShiftChecker interface {
	Covers(ctx context.Context, trainerID int64, class Interval) (bool, error)
}

// bookingStore คือสิ่งที่ Service ต้องใช้จาก Store (CON-DB-01).
type bookingStore interface {
	ListUpcoming(ctx context.Context, now time.Time) ([]Class, error)
	GetForMember(ctx context.Context, bookingID, memberID int64) (Booking, error)
	Book(ctx context.Context, classID, memberID int64, now time.Time, checkShift ShiftCheck) (Booking, error)
}

// noShiftData คือ ShiftChecker เมื่อยังไม่มีแหล่งข้อมูลกะ: ไม่มีข้อมูลกะ = นอกกะ = บล็อก (DOM-BKG-01, ASM-08, plan D6).
type noShiftData struct{}

// Covers ตอบว่าไม่อยู่ในกะเสมอ.
func (noShiftData) Covers(context.Context, int64, Interval) (bool, error) { return false, nil }

// Service คือกฎการจองคลาสที่ handler เรียกใช้ (FR-BKG-01, FR-BKG-02, DOM-BKG-01).
type Service struct {
	store  bookingStore
	shifts ShiftChecker
	clock  Clock
}

// NewService ประกอบ Service; shifts เป็น nil หมายถึงยังไม่มีข้อมูลกะ (ทุกคลาสนอกกะ) และ clock เป็น nil ใช้ SystemClock
// (FR-BKG-02, ASM-08, plan D6).
func NewService(store bookingStore, shifts ShiftChecker, clock Clock) *Service {
	if shifts == nil {
		shifts = noShiftData{}
	}
	if clock == nil {
		clock = SystemClock{}
	}
	return &Service{store: store, shifts: shifts, clock: clock}
}

// Classes คืนคลาสที่ยังไม่ถึงเวลาเริ่มให้สมาชิกเลือก (FR-BKG-01, plan D4).
func (s *Service) Classes(ctx context.Context) ([]Class, error) {
	return s.store.ListUpcoming(ctx, s.clock.Now())
}

// Book จองคลาสให้สมาชิกที่ล็อกอินอยู่ ตรวจทุกเงื่อนไขใน transaction ของ Store (FR-BKG-01, FR-BKG-02, NFR-PERF-01).
func (s *Service) Book(ctx context.Context, classID, memberID int64) (Booking, error) {
	return s.store.Book(ctx, classID, memberID, s.clock.Now(), s.checkShift)
}

// Booking คืนการจองของสมาชิกทีละรายการ (FR-BKG-01 AC-BKG-01, plan D5).
func (s *Service) Booking(ctx context.Context, bookingID, memberID int64) (Booking, error) {
	return s.store.GetForMember(ctx, bookingID, memberID)
}

// checkShift แปลงผลของ ShiftChecker เป็น ShiftCheck: นอกกะได้ ErrTrainerOffShift, อ่านไม่ได้ได้ error ที่ไม่ใช่เหตุผลการจอง
// (handler ตอบ 503 ตาม ASM-07) (DOM-BKG-01, ASM-08, ASM-11).
func (s *Service) checkShift(ctx context.Context, trainerID int64, class Interval) error {
	ok, err := s.shifts.Covers(ctx, trainerID, class)
	if err != nil {
		return fmt.Errorf("booking: read trainer shift: %w", err)
	}
	if !ok {
		return ErrTrainerOffShift
	}
	return nil
}
