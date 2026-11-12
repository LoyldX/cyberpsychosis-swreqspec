package roster

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Trainer คือเทรนเนอร์ที่เตรียมไว้ล่วงหน้า (FR-ROST-01, ASM-09).
type Trainer struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Entry คือรายการตารางงานหนึ่งรายการ: เวลาเข้างานหรือเวลาสอนซ้ำทุกสัปดาห์ (FR-ROST-01, ASM-01, ASM-03, ASM-04).
type Entry struct {
	ID        int64
	TrainerID int64
	Kind      Kind
	Slot      Slot
	Cancelled bool
}

// MarshalJSON แสดงรายการตามสัญญา plan ข้อ 4: {id, trainer_id, kind, weekday, start_time, end_time, cancelled}.
func (e Entry) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID        int64  `json:"id"`
		TrainerID int64  `json:"trainer_id"`
		Kind      Kind   `json:"kind"`
		Weekday   int    `json:"weekday"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Cancelled bool   `json:"cancelled"`
	}{e.ID, e.TrainerID, e.Kind, e.Slot.Weekday, e.Slot.Start.String(), e.Slot.End.String(), e.Cancelled})
}

// trainerLockNamespace คือ key แรกของ advisory lock แบบ 2 key ต่อเทรนเนอร์ ใช้ค่าเดียวกับการจองของ 005
// การแก้ตารางกับการจองของเทรนเนอร์คนเดียวกันจึงผ่านทีละรายการ (ASM-07, plan D2).
const trainerLockNamespace = 5_005

// entryColumns คือคอลัมน์ที่อ่าน Entry โดยแปลง TIME เป็นวินาทีในวัน.
const entryColumns = `id, trainer_id, kind, weekday,
	EXTRACT(EPOCH FROM start_time)::int, EXTRACT(EPOCH FROM end_time)::int, cancelled`

// Store อ่าน/เขียนตารางงานใน PostgreSQL (CON-DB-01).
type Store struct {
	pool *pgxpool.Pool
}

// NewStore สร้าง Store จาก pool ที่ใช้ร่วมกับทั้งแอป (CON-DB-01).
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// ListTrainers คืนเทรนเนอร์ทั้งหมดเรียงตามรหัส (FR-ROST-01, ASM-09).
func (s *Store) ListTrainers(ctx context.Context) ([]Trainer, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name FROM trainers ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("roster: list trainers: %w", err)
	}
	trainers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[Trainer])
	if err != nil {
		return nil, fmt.Errorf("roster: list trainers: %w", err)
	}
	return trainers, nil
}

// ListEntries คืนรายการทั้งหมดของเทรนเนอร์ (รวมที่ยกเลิก plan D5) เรียงตามวันและเวลาเริ่ม
// เทรนเนอร์ไม่มีอยู่คืน ErrTrainerNotFound (FR-ROST-01, AC-ROST-01..04 "เรียกดูได้").
func (s *Store) ListEntries(ctx context.Context, trainerID int64) ([]Entry, error) {
	if err := trainerExists(ctx, s.pool, trainerID); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx,
		`SELECT `+entryColumns+` FROM roster_entries WHERE trainer_id = $1 ORDER BY weekday, start_time, id`, trainerID)
	if err != nil {
		return nil, fmt.Errorf("roster: list entries: %w", err)
	}
	return collectEntries(rows)
}

// Create เพิ่มรายการใหม่ให้เทรนเนอร์ใน transaction เดียวพร้อมล็อกเทรนเนอร์
// ตรวจรูปแบบ → ไม่พบเทรนเนอร์ → ทับซ้อน → เวลาสอนนอกเวลาเข้างาน ไม่ผ่านแล้วไม่บันทึก
// (FR-ROST-01, AC-ROST-01, AC-ROST-02, CON-DB-01; ASM-02, ASM-03, ASM-04, ASM-10).
func (s *Store) Create(ctx context.Context, trainerID int64, kind Kind, slot Slot) (Entry, error) {
	if _, err := ParseKind(string(kind)); err != nil {
		return Entry{}, err
	}
	if err := slot.Validate(); err != nil {
		return Entry{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Entry{}, fmt.Errorf("roster: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // ไม่มีผลหลัง Commit; ทุกทางที่ไม่ผ่านจึงไม่บันทึก (ASM-02, ASM-08)

	if err := trainerExists(ctx, tx, trainerID); err != nil {
		return Entry{}, err
	}
	if err := lockTrainer(ctx, tx, trainerID); err != nil {
		return Entry{}, err
	}
	active, err := activeEntries(ctx, tx, trainerID, slot.Weekday)
	if err != nil {
		return Entry{}, err
	}
	candidate := Entry{TrainerID: trainerID, Kind: kind, Slot: slot}
	if err := checkSchedule(append(active, candidate)); err != nil {
		return Entry{}, err
	}

	row := tx.QueryRow(ctx,
		`INSERT INTO roster_entries (trainer_id, kind, weekday, start_time, end_time)
		 VALUES ($1, $2, $3, ($4::int * interval '1 second')::time, ($5::int * interval '1 second')::time)
		 RETURNING `+entryColumns,
		trainerID, string(kind), slot.Weekday, int(slot.Start), int(slot.End))
	created, err := scanEntry(row)
	if err != nil {
		return Entry{}, fmt.Errorf("roster: insert entry: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Entry{}, fmt.Errorf("roster: commit: %w", err)
	}
	return created, nil
}

// checkSchedule ตรวจรายการที่ใช้งาน (ไม่ถูกยกเลิก) ของเทรนเนอร์ในวันเดียวกันหลังการเปลี่ยนแปลง:
// ประเภทเดียวกันห้ามทับซ้อน (ต่อกันพอดีได้) แล้วเวลาสอนทุกรายการต้องอยู่ในเวลาเข้างาน (ASM-02, ASM-04, ASM-10).
func checkSchedule(entries []Entry) error {
	var work []Slot
	for i, a := range entries {
		for _, b := range entries[i+1:] {
			if a.Kind == b.Kind && a.Slot.Overlaps(b.Slot) {
				return ErrOverlap
			}
		}
		if a.Kind == KindWork {
			work = append(work, a.Slot)
		}
	}
	for _, e := range entries {
		if e.Kind == KindTeach && !e.Slot.Within(work) {
			return ErrTeachOutsideWork
		}
	}
	return nil
}

// querier คือสิ่งที่รัน query ได้ทั้ง pool และ transaction.
type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// trainerExists คืน ErrTrainerNotFound ถ้าไม่มีเทรนเนอร์ (ASM-09).
func trainerExists(ctx context.Context, q querier, trainerID int64) error {
	var exists bool
	if err := q.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM trainers WHERE id = $1)`, trainerID).Scan(&exists); err != nil {
		return fmt.Errorf("roster: check trainer: %w", err)
	}
	if !exists {
		return ErrTrainerNotFound
	}
	return nil
}

// lockTrainer ล็อกเทรนเนอร์จนจบ transaction ด้วย key เดียวกับการจองของ 005 (ASM-07, plan D2).
func lockTrainer(ctx context.Context, tx pgx.Tx, trainerID int64) error {
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1, ($2::bigint % 2147483647)::int)`,
		trainerLockNamespace, trainerID); err != nil {
		return fmt.Errorf("roster: lock trainer: %w", err)
	}
	return nil
}

// activeEntries คืนรายการที่ไม่ถูกยกเลิกของเทรนเนอร์ในวันหนึ่ง (ASM-01: ยกเลิกแล้วไม่นับ).
func activeEntries(ctx context.Context, q querier, trainerID int64, weekday int) ([]Entry, error) {
	rows, err := q.Query(ctx,
		`SELECT `+entryColumns+` FROM roster_entries
		 WHERE trainer_id = $1 AND weekday = $2 AND NOT cancelled`, trainerID, weekday)
	if err != nil {
		return nil, fmt.Errorf("roster: read entries: %w", err)
	}
	return collectEntries(rows)
}

// collectEntries อ่านทุกแถวเป็น Entry.
func collectEntries(rows pgx.Rows) ([]Entry, error) {
	entries, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Entry, error) { return scanEntry(row) })
	if err != nil {
		return nil, fmt.Errorf("roster: read entries: %w", err)
	}
	return entries, nil
}

// scanEntry อ่านหนึ่งแถวตาม entryColumns.
func scanEntry(row pgx.Row) (Entry, error) {
	var e Entry
	var kind string
	var start, end int
	if err := row.Scan(&e.ID, &e.TrainerID, &kind, &e.Slot.Weekday, &start, &end, &e.Cancelled); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Entry{}, ErrEntryNotFound
		}
		return Entry{}, err
	}
	e.Kind, e.Slot.Start, e.Slot.End = Kind(kind), TimeOfDay(start), TimeOfDay(end)
	return e, nil
}
