package roster

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// change คือผลของการแก้รายการหนึ่งรายการ: รายการใหม่ (nil = ลบ).
type change func(old Entry) (*Entry, error)

// Update แก้วันและเวลาของรายการทีละรายการ (เปลี่ยนประเภทไม่ได้) (FR-ROST-01, AC-ROST-03, AC-ROST-04; ASM-05).
func (s *Store) Update(ctx context.Context, entryID int64, slot Slot, now time.Time) (Entry, error) {
	if err := slot.Validate(); err != nil {
		return Entry{}, err
	}
	return s.apply(ctx, entryID, now, func(old Entry) (*Entry, error) {
		updated := old
		updated.Slot = slot
		return &updated, nil
	})
}

// Cancel ยกเลิกรายการ: ยังเก็บไว้แต่ไม่นับเป็นเวลาเข้างาน/เวลาสอน ยกเลิกซ้ำได้ผลเดิม (FR-ROST-01; ASM-01).
func (s *Store) Cancel(ctx context.Context, entryID int64, now time.Time) (Entry, error) {
	return s.apply(ctx, entryID, now, func(old Entry) (*Entry, error) {
		cancelled := old
		cancelled.Cancelled = true
		return &cancelled, nil
	})
}

// Delete ลบรายการออกจากระบบ (FR-ROST-01; ASM-01).
func (s *Store) Delete(ctx context.Context, entryID int64, now time.Time) error {
	_, err := s.apply(ctx, entryID, now, func(Entry) (*Entry, error) { return nil, nil })
	return err
}

// apply ทำการเปลี่ยนแปลงใน transaction เดียวพร้อมล็อกเทรนเนอร์ แล้วตรวจตามลำดับ plan ข้อ 4:
// ไม่พบ → ทับซ้อน → เวลาสอนนอกเวลาเข้างาน → คลาสที่มีผู้จองและยังไม่เริ่มหลุดเวลาเข้างาน ไม่ผ่านแล้วไม่บันทึก
// (FR-ROST-01, CON-DB-01; ASM-02, ASM-04, ASM-07, ASM-10, Q-11, plan D2).
func (s *Store) apply(ctx context.Context, entryID int64, now time.Time, fn change) (Entry, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Entry{}, fmt.Errorf("roster: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // ไม่มีผลหลัง Commit; ทุกทางที่ไม่ผ่านจึงไม่บันทึก (ASM-02, ASM-08)

	old, err := scanEntry(tx.QueryRow(ctx, `SELECT `+entryColumns+` FROM roster_entries WHERE id = $1`, entryID))
	if err != nil {
		return Entry{}, wrapRead(err)
	}
	if err := lockTrainer(ctx, tx, old.TrainerID); err != nil {
		return Entry{}, err
	}
	// อ่านซ้ำหลังล็อก เผื่อมีคำขออื่นแก้/ลบไปก่อน
	old, err = scanEntry(tx.QueryRow(ctx, `SELECT `+entryColumns+` FROM roster_entries WHERE id = $1`, entryID))
	if err != nil {
		return Entry{}, wrapRead(err)
	}
	next, err := fn(old)
	if err != nil {
		return Entry{}, err
	}

	before, err := trainerActiveEntries(ctx, tx, old.TrainerID)
	if err != nil {
		return Entry{}, err
	}
	after := replaceEntry(before, old.ID, next)

	for _, weekday := range affectedWeekdays(old, next) {
		if err := checkSchedule(onWeekday(after, weekday)); err != nil {
			return Entry{}, err
		}
	}
	if old.Kind == KindWork {
		if err := checkBookedClasses(ctx, tx, old.TrainerID, now, workSlots(before), workSlots(after)); err != nil {
			return Entry{}, err
		}
	}

	result, err := writeChange(ctx, tx, old, next)
	if err != nil {
		return Entry{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Entry{}, fmt.Errorf("roster: commit: %w", err)
	}
	return result, nil
}

// wrapRead คง ErrEntryNotFound ไว้ และห่อ error อื่นของฐานข้อมูล.
func wrapRead(err error) error {
	if err == ErrEntryNotFound {
		return err
	}
	return fmt.Errorf("roster: read entry: %w", err)
}

// trainerActiveEntries คืนรายการที่ไม่ถูกยกเลิกทั้งหมดของเทรนเนอร์ทุกวัน (ASM-01).
func trainerActiveEntries(ctx context.Context, tx pgx.Tx, trainerID int64) ([]Entry, error) {
	rows, err := tx.Query(ctx,
		`SELECT `+entryColumns+` FROM roster_entries WHERE trainer_id = $1 AND NOT cancelled`, trainerID)
	if err != nil {
		return nil, fmt.Errorf("roster: read entries: %w", err)
	}
	return collectEntries(rows)
}

// replaceEntry คืนรายการที่ใช้งานหลังการเปลี่ยนแปลง: เอารายการเดิมออก แล้วใส่รายการใหม่ถ้ายังใช้งาน (ASM-01).
func replaceEntry(active []Entry, id int64, next *Entry) []Entry {
	out := make([]Entry, 0, len(active)+1)
	for _, e := range active {
		if e.ID != id {
			out = append(out, e)
		}
	}
	if next != nil && !next.Cancelled {
		out = append(out, *next)
	}
	return out
}

// affectedWeekdays คือวันที่ต้องตรวจซ้ำ: วันเดิมและวันใหม่ (ถ้าย้ายวัน).
func affectedWeekdays(old Entry, next *Entry) []int {
	days := []int{old.Slot.Weekday}
	if next != nil && next.Slot.Weekday != old.Slot.Weekday {
		days = append([]int{next.Slot.Weekday}, days...)
	}
	return days
}

// onWeekday กรองรายการของวันหนึ่ง.
func onWeekday(entries []Entry, weekday int) []Entry {
	var out []Entry
	for _, e := range entries {
		if e.Slot.Weekday == weekday {
			out = append(out, e)
		}
	}
	return out
}

// workSlots คืนช่วงเวลาเข้างานจากรายการที่ใช้งาน (ASM-04).
func workSlots(entries []Entry) []Slot {
	var out []Slot
	for _, e := range entries {
		if e.Kind == KindWork {
			out = append(out, e.Slot)
		}
	}
	return out
}

// checkBookedClasses ปฏิเสธถ้าคลาสของเทรนเนอร์ที่มีผู้จองแล้วและยังไม่ถึงเวลาเริ่ม ซึ่งอยู่ในเวลาเข้างานก่อนการเปลี่ยนแปลง
// จะไม่อยู่ในเวลาเข้างานอีกต่อไปหลังการเปลี่ยนแปลง (ASM-07 "ไม่อยู่ในเวลาเข้างาน...อีกต่อไป", Q-11; ตาราง classes/bookings ของ 005).
func checkBookedClasses(ctx context.Context, tx pgx.Tx, trainerID int64, now time.Time, before, after []Slot) error {
	rows, err := tx.Query(ctx,
		`SELECT c.starts_at, c.ends_at FROM classes c
		 WHERE c.trainer_id = $1 AND c.starts_at > $2
		   AND EXISTS (SELECT 1 FROM bookings b WHERE b.class_id = c.id)`, trainerID, now)
	if err != nil {
		return fmt.Errorf("roster: read booked classes: %w", err)
	}
	classes, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) ([2]time.Time, error) {
		var c [2]time.Time
		err := row.Scan(&c[0], &c[1])
		return c, err
	})
	if err != nil {
		return fmt.Errorf("roster: read booked classes: %w", err)
	}
	for _, c := range classes {
		slot, ok := SlotOf(c[0], c[1])
		if ok && slot.Within(before) && !slot.Within(after) {
			return ErrBookedClassOutsideWork
		}
	}
	return nil
}

// writeChange บันทึกการเปลี่ยนแปลงลงฐานข้อมูล แล้วคืนรายการผลลัพธ์ (ลบแล้วคืนรายการเดิม).
func writeChange(ctx context.Context, tx pgx.Tx, old Entry, next *Entry) (Entry, error) {
	if next == nil {
		if _, err := tx.Exec(ctx, `DELETE FROM roster_entries WHERE id = $1`, old.ID); err != nil {
			return Entry{}, fmt.Errorf("roster: delete entry: %w", err)
		}
		return old, nil
	}
	row := tx.QueryRow(ctx,
		`UPDATE roster_entries
		 SET weekday = $2, start_time = ($3::int * interval '1 second')::time,
		     end_time = ($4::int * interval '1 second')::time, cancelled = $5
		 WHERE id = $1 RETURNING `+entryColumns,
		old.ID, next.Slot.Weekday, int(next.Slot.Start), int(next.Slot.End), next.Cancelled)
	updated, err := scanEntry(row)
	if err != nil {
		return Entry{}, fmt.Errorf("roster: update entry: %w", err)
	}
	return updated, nil
}
