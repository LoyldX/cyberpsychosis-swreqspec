package booking

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Class คือคลาสกลุ่มที่เตรียมไว้ล่วงหน้า (FR-BKG-01, ASM-03, ASM-04).
type Class struct {
	ID        int64     `json:"id"`
	TrainerID int64     `json:"trainer_id"`
	StartsAt  time.Time `json:"starts_at"`
	EndsAt    time.Time `json:"ends_at"`
	Capacity  int       `json:"capacity"`
}

// Booking คือการจองคลาสของสมาชิกหนึ่งคน (FR-BKG-01, ASM-06).
type Booking struct {
	ID        int64     `json:"id"`
	ClassID   int64     `json:"class_id"`
	MemberID  int64     `json:"member_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ShiftCheck ตรวจว่าช่วงเวลาของคลาสอยู่ในเวลาเข้างานของเทรนเนอร์หรือไม่ คืน nil ถ้าอยู่ในกะ
// คืน ErrTrainerOffShift ถ้านอกกะ หรือ error อื่นถ้าอ่านข้อมูลไม่ได้ (FR-BKG-02, DOM-BKG-01, ASM-08, ASM-11).
type ShiftCheck func(ctx context.Context, trainerID int64, class Interval) error

// bookingLockNamespace คือ key แรกของ advisory lock แบบ 2 key ต่อเทรนเนอร์ (plan D2)
// key space แบบ 2 key ไม่ทับกับแบบ key เดียวที่ 007 ใช้ (7001).
const bookingLockNamespace = 5_005

// Store อ่าน/เขียนคลาสและการจองใน PostgreSQL (CON-DB-01).
type Store struct {
	pool *pgxpool.Pool
}

// NewStore สร้าง Store จาก pool ที่ใช้ร่วมกับทั้งแอป (CON-DB-01).
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// ListUpcoming คืนคลาสที่ยังไม่ถึงเวลาเริ่ม ณ now เรียงตามเวลาเริ่ม (FR-BKG-01, ASM-05, plan D4).
func (s *Store) ListUpcoming(ctx context.Context, now time.Time) ([]Class, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, trainer_id, starts_at, ends_at, capacity FROM classes
		 WHERE starts_at > $1 ORDER BY starts_at, id`, now)
	if err != nil {
		return nil, fmt.Errorf("booking: list classes: %w", err)
	}
	classes, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Class, error) {
		var c Class
		err := row.Scan(&c.ID, &c.TrainerID, &c.StartsAt, &c.EndsAt, &c.Capacity)
		return c, err
	})
	if err != nil {
		return nil, fmt.Errorf("booking: list classes: %w", err)
	}
	return classes, nil
}

// GetForMember คืนการจอง bookingID ถ้าเป็นของ memberID ไม่เช่นนั้นคืน ErrBookingNotFound
// (FR-BKG-01 AC-BKG-01 "เรียกดูการจองนั้นได้", plan D5).
func (s *Store) GetForMember(ctx context.Context, bookingID, memberID int64) (Booking, error) {
	var b Booking
	err := s.pool.QueryRow(ctx,
		`SELECT id, class_id, member_id, created_at FROM bookings WHERE id = $1 AND member_id = $2`,
		bookingID, memberID).Scan(&b.ID, &b.ClassID, &b.MemberID, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Booking{}, ErrBookingNotFound
	}
	if err != nil {
		return Booking{}, fmt.Errorf("booking: get booking: %w", err)
	}
	return b, nil
}

// Book จองคลาส classID ให้ memberID ใน transaction เดียว ล็อกทีละเทรนเนอร์ แล้วตรวจตามลำดับ plan D3:
// ไม่พบคลาส → เริ่มแล้ว → จองซ้ำ → เต็ม → นอกกะ → ทับซ้อน ถ้าไม่ผ่านข้อใดจะไม่บันทึกอะไรเลย
// (FR-BKG-01, FR-BKG-02, DOM-BKG-01, NFR-PERF-01, CON-DB-01; ASM-03, ASM-05, ASM-07, ASM-10).
func (s *Store) Book(ctx context.Context, classID, memberID int64, now time.Time, checkShift ShiftCheck) (Booking, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Booking{}, fmt.Errorf("booking: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // ไม่มีผลหลัง Commit; ทุกทางที่ไม่ผ่านจึงไม่บันทึก (ASM-07)

	var trainerID int64
	err = tx.QueryRow(ctx, `SELECT trainer_id FROM classes WHERE id = $1`, classID).Scan(&trainerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Booking{}, ErrClassNotFound
	}
	if err != nil {
		return Booking{}, fmt.Errorf("booking: read class: %w", err)
	}

	// คำขอของเทรนเนอร์คนเดียวกันผ่านทีละรายการ จนกว่า transaction นี้จบ (NFR-PERF-01, AC-BKG-05, plan D2)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1, ($2::bigint % 2147483647)::int)`,
		bookingLockNamespace, trainerID); err != nil {
		return Booking{}, fmt.Errorf("booking: lock trainer: %w", err)
	}

	var class Class
	err = tx.QueryRow(ctx,
		`SELECT id, trainer_id, starts_at, ends_at, capacity FROM classes WHERE id = $1`, classID).
		Scan(&class.ID, &class.TrainerID, &class.StartsAt, &class.EndsAt, &class.Capacity)
	if errors.Is(err, pgx.ErrNoRows) {
		return Booking{}, ErrClassNotFound
	}
	if err != nil {
		return Booking{}, fmt.Errorf("booking: read class: %w", err)
	}
	interval := Interval{Start: class.StartsAt, End: class.EndsAt}

	if !now.Before(class.StartsAt) { // ถึงเวลาเริ่มแล้วจองไม่ได้ (ASM-05, Q-12)
		return Booking{}, ErrClassStarted
	}

	var alreadyBooked bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM bookings WHERE class_id = $1 AND member_id = $2)`,
		classID, memberID).Scan(&alreadyBooked); err != nil {
		return Booking{}, fmt.Errorf("booking: check duplicate: %w", err)
	}
	if alreadyBooked { // ASM-05
		return Booking{}, ErrAlreadyBooked
	}

	var booked int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM bookings WHERE class_id = $1`, classID).Scan(&booked); err != nil {
		return Booking{}, fmt.Errorf("booking: count seats: %w", err)
	}
	if booked >= class.Capacity { // คลาสกลุ่มมีที่นั่งจำกัด (ASM-03, ASM-05)
		return Booking{}, ErrClassFull
	}

	if err := checkShift(ctx, class.TrainerID, interval); err != nil { // นอกกะ (DOM-BKG-01, ASM-08, ASM-11)
		return Booking{}, err
	}

	busy, err := trainerBusy(ctx, tx, class, interval)
	if err != nil {
		return Booking{}, err
	}
	if busy { // ทับซ้อนกับคิวสอนอื่น (DOM-BKG-01, ASM-09, ASM-10, ASM-12)
		return Booking{}, ErrTrainerBusy
	}

	var b Booking
	err = tx.QueryRow(ctx,
		`INSERT INTO bookings (class_id, member_id) VALUES ($1, $2) RETURNING id, class_id, member_id, created_at`,
		classID, memberID).Scan(&b.ID, &b.ClassID, &b.MemberID, &b.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // กันซ้ำชั้นสุดท้ายด้วย UNIQUE (ASM-05)
		return Booking{}, ErrAlreadyBooked
	}
	if err != nil {
		return Booking{}, fmt.Errorf("booking: insert: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Booking{}, fmt.Errorf("booking: commit: %w", err)
	}
	return b, nil
}

// trainerBusy บอกว่าเทรนเนอร์ของ class มีคลาสอื่นที่มีผู้จองแล้ว (คิวสอน ASM-10) และเวลาทับกับ interval หรือไม่
// คิวสอนอื่น = คลาสอื่นของเทรนเนอร์คนเดียวกันเท่านั้น (ASM-09) เทียบด้วย Interval.Overlaps (ASM-12).
func trainerBusy(ctx context.Context, tx pgx.Tx, class Class, interval Interval) (bool, error) {
	rows, err := tx.Query(ctx,
		`SELECT c.starts_at, c.ends_at FROM classes c
		 WHERE c.trainer_id = $1 AND c.id <> $2 AND c.starts_at <= $4 AND c.ends_at >= $3
		   AND EXISTS (SELECT 1 FROM bookings b WHERE b.class_id = c.id)`,
		class.TrainerID, class.ID, interval.Start, interval.End)
	if err != nil {
		return false, fmt.Errorf("booking: check trainer queue: %w", err)
	}
	others, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Interval, error) {
		var iv Interval
		err := row.Scan(&iv.Start, &iv.End)
		return iv, err
	})
	if err != nil {
		return false, fmt.Errorf("booking: check trainer queue: %w", err)
	}
	for _, other := range others {
		if interval.Overlaps(other) {
			return true, nil
		}
	}
	return false, nil
}
