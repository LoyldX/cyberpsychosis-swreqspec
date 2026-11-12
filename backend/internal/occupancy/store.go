package occupancy

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// gateSignalsLockKey คือ key ของ pg_advisory_xact_lock ที่ทำให้สัญญาณถูกคำนวณทีละตัวตามลำดับที่ได้รับ
// (ASM-05, plan D2) — ค่าตัวเลขทีมเลือกเอง ไม่ได้มาจาก spec ขอแค่ไม่ซ้ำกับ advisory lock อื่นในระบบ.
const gateSignalsLockKey int64 = 7_001

// Store บันทึกสัญญาณผ่านประตูและอ่านจำนวนปัจจุบันจาก PostgreSQL (FR-BI-01, DOM-CAP-01, IF-GATE-01; ASM-02).
type Store struct {
	pool *pgxpool.Pool
}

// NewStore สร้าง Store บน pool ที่ผู้เรียกเป็นเจ้าของ (ASM-02).
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// rowQuerier คือส่วนที่ใช้ร่วมกันของ *pgxpool.Pool และ pgx.Tx.
type rowQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// RecordSignal บันทึกสัญญาณ 1 ครั้งที่ได้รับเมื่อเวลา at แล้วคืนจำนวนปัจจุบันหลังสัญญาณนี้
// ทำใน transaction เดียวภายใต้ advisory lock: อ่านจำนวนของวันนั้น → ApplySignal → insert พร้อม occupancy_after
// (DOM-CAP-01, ASM-03, ASM-04, ASM-05: รับทุกสัญญาณ ไม่กรองซ้ำ และเก็บสัญญาณออกที่จำนวนเป็น 0 ด้วย).
func (s *Store) RecordSignal(ctx context.Context, d Direction, at time.Time) (int, error) {
	if _, err := ParseDirection(string(d)); err != nil {
		return 0, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("occupancy: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, gateSignalsLockKey); err != nil {
		return 0, fmt.Errorf("occupancy: lock: %w", err)
	}
	current, err := countSince(ctx, tx, StartOfDay(at))
	if err != nil {
		return 0, err
	}
	next := ApplySignal(current, d)
	if _, err := tx.Exec(ctx, `
		INSERT INTO gate_signals (direction, received_at, occupancy_after)
		VALUES ($1, $2, $3)`, string(d), at, next); err != nil {
		return 0, fmt.Errorf("occupancy: insert signal: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("occupancy: commit: %w", err)
	}
	return next, nil
}

// CurrentCount คืนจำนวนปัจจุบัน ณ เวลา now = occupancy_after ของสัญญาณล่าสุดตั้งแต่ 00:00 Asia/Bangkok ของวันนั้น
// ถ้าวันนั้นยังไม่มีสัญญาณ คืน 0 (FR-BI-01, DOM-CAP-01, ASM-02, ASM-03).
func (s *Store) CurrentCount(ctx context.Context, now time.Time) (int, error) {
	return countSince(ctx, s.pool, StartOfDay(now))
}

// countSince อ่าน occupancy_after ของแถวล่าสุด (เรียงตามลำดับที่ได้รับ) ที่ received_at >= since (ASM-03, ASM-05).
func countSince(ctx context.Context, q rowQuerier, since time.Time) (int, error) {
	var count int
	err := q.QueryRow(ctx, `
		SELECT occupancy_after
		FROM gate_signals
		WHERE received_at >= $1
		ORDER BY id DESC
		LIMIT 1`, since).Scan(&count)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("occupancy: current count: %w", err)
	}
	return count, nil
}
