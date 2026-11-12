package booking

import (
	"context"
	"time"
)

// rosterCoverer คือสิ่งที่ 004 ส่งมอบ: ตอบว่าช่วงเวลาอยู่ในเวลาเข้างานของเทรนเนอร์หรือไม่ (*roster.Store ตรงตามนี้)
// (FR-BKG-02, DOM-BKG-01; 005 ASM-11, ASM-13; 004 ASM-03, ASM-10).
type rosterCoverer interface {
	Covers(ctx context.Context, trainerID int64, start, end time.Time) (bool, error)
}

// RosterShifts คือ ShiftChecker ตัวจริงที่อ่านเวลาเข้างานจากตารางงานของ 004 (FR-BKG-02, DOM-BKG-01, ASM-11, plan D6).
type RosterShifts struct {
	roster rosterCoverer
}

// NewRosterShifts ห่อตัวตอบเวลาเข้างานของ 004 ให้เป็น ShiftChecker (ASM-11).
func NewRosterShifts(roster rosterCoverer) *RosterShifts {
	return &RosterShifts{roster: roster}
}

// Covers ตอบว่าคลาสอยู่ในเวลาเข้างานของเทรนเนอร์ทั้งหมดหรือไม่ ขอบพอดีนับว่าอยู่ในกะ (DOM-BKG-01, ASM-08, ASM-13).
func (s *RosterShifts) Covers(ctx context.Context, trainerID int64, class Interval) (bool, error) {
	return s.roster.Covers(ctx, trainerID, class.Start, class.End)
}
