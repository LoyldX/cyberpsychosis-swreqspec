package occupancy

import (
	"context"
	"log"
	"sync"
	"time"
)

// Clock ให้เวลาปัจจุบันและตัวจับเวลา แยกออกมาเพื่อให้ test จำลองเวลา 00:00 ได้ (ASM-03, AC-BI-06).
type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

// SystemClock คือ Clock ที่ใช้เวลาจริงของเครื่อง.
type SystemClock struct{}

// Now คืนเวลาปัจจุบันจริง.
func (SystemClock) Now() time.Time { return time.Now() }

// After คืนช่องที่ได้ค่าเมื่อเวลาผ่านไป d.
func (SystemClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// counter คือส่วนของ Store ที่ Service ใช้ (FR-BI-01, DOM-CAP-01).
type counter interface {
	RecordSignal(ctx context.Context, d Direction, at time.Time) (int, error)
	CurrentCount(ctx context.Context, now time.Time) (int, error)
}

// Service ประกอบ Store + Broadcaster + Clock: บันทึกสัญญาณแล้วกระจายจำนวนใหม่ทันที
// และกระจายจำนวนใหม่เมื่อถึง 00:00 Asia/Bangkok (FR-BI-01, NFR-PERF-02, DOM-CAP-01; ASM-03, ASM-06, plan D3).
type Service struct {
	store counter
	hub   *Broadcaster
	clock Clock
	// mu ทำให้ลำดับการ Publish ตรงกับลำดับการบันทึก ภายใน process เดียว (ASM-05).
	mu sync.Mutex
}

// NewService สร้าง Service; clock เป็น nil ได้ (ใช้ SystemClock).
func NewService(store counter, hub *Broadcaster, clock Clock) *Service {
	if clock == nil {
		clock = SystemClock{}
	}
	return &Service{store: store, hub: hub, clock: clock}
}

// Record บันทึกสัญญาณ 1 ครั้ง ณ เวลาปัจจุบันของ clock แล้วกระจายจำนวนใหม่ให้ผู้ฟังทุกราย (DOM-CAP-01, NFR-PERF-02, IF-GATE-01).
func (s *Service) Record(ctx context.Context, d Direction) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.clock.Now()
	count, err := s.store.RecordSignal(ctx, d, now)
	if err != nil {
		return Snapshot{}, err
	}
	snap := Snapshot{Count: count, AsOf: now}
	s.hub.Publish(snap)
	return snap, nil
}

// Current คืนจำนวนปัจจุบัน ณ เวลาปัจจุบันของ clock (FR-BI-01, ASM-03).
func (s *Service) Current(ctx context.Context) (Snapshot, error) {
	now := s.clock.Now()
	count, err := s.store.CurrentCount(ctx, now)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Count: count, AsOf: now}, nil
}

// Subscribe ลงทะเบียนผู้ฟังจำนวนใหม่ (NFR-PERF-02).
func (s *Service) Subscribe() (<-chan Snapshot, func()) {
	return s.hub.Subscribe()
}

// RunMidnightReset รอจนถึง 00:00 Asia/Bangkok ของวันถัดไปซ้ำ ๆ แล้วกระจายจำนวนปัจจุบันของวันใหม่ (เป็น 0 ถ้ายังไม่มีสัญญาณ)
// ทำงานจน ctx ถูกยกเลิก (ASM-03, AC-BI-06, plan D3).
func (s *Service) RunMidnightReset(ctx context.Context) {
	for {
		now := s.clock.Now()
		wait := NextStartOfDay(now).Sub(now)
		select {
		case <-ctx.Done():
			return
		case <-s.clock.After(wait):
		}
		s.mu.Lock()
		snap, err := s.Current(ctx)
		if err == nil {
			s.hub.Publish(snap)
		}
		s.mu.Unlock()
		if err != nil && ctx.Err() == nil {
			log.Printf("occupancy: midnight reset: %v", err)
		}
	}
}
