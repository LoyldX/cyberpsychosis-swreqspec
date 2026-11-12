package roster

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gymapp/backend/internal/testutil"
)

// newTestPool ต่อฐานข้อมูลทดสอบ (migrate ถึง 0005) และล้างตารางของ 004 ก่อนและหลัง test
// (plan D8: ฐานข้อมูลเดียวกับ dev; ไม่ใช้ t.Parallel).
func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := testutil.ConnectPool(t)
	resetRosterTables(t, pool)
	t.Cleanup(func() { resetRosterTables(t, pool) })
	return pool
}

// resetRosterTables ลบรายการตารางงานและเทรนเนอร์ทั้งหมด (ข้อมูลทดสอบของ 004 เท่านั้น).
func resetRosterTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	_, err := pool.Exec(ctx, `DELETE FROM roster_entries`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `DELETE FROM trainers`)
	require.NoError(t, err)
}

// newTrainer เพิ่มเทรนเนอร์ที่เตรียมไว้ล่วงหน้า (ASM-09) แล้วคืน id.
func newTrainer(t *testing.T, pool *pgxpool.Pool, name string) int64 {
	t.Helper()
	var id int64
	require.NoError(t, pool.QueryRow(context.Background(), `INSERT INTO trainers (name) VALUES ($1) RETURNING id`, name).Scan(&id))
	return id
}

// countEntries นับรายการของเทรนเนอร์ ใช้ยืนยันว่า "ไม่บันทึก" (ASM-02).
func countEntries(t *testing.T, pool *pgxpool.Pool, trainerID int64) int {
	t.Helper()
	var n int
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT count(*) FROM roster_entries WHERE trainer_id = $1`, trainerID).Scan(&n))
	return n
}

// mustCreate เพิ่มรายการที่ต้องสำเร็จ.
func mustCreate(t *testing.T, store *Store, trainerID int64, kind Kind, s Slot) Entry {
	t.Helper()
	e, err := store.Create(context.Background(), trainerID, kind, s)
	require.NoError(t, err)
	return e
}

// TestStoreListTrainers: คืนเทรนเนอร์ทั้งหมดเรียงตามรหัส (ASM-09).
func TestStoreListTrainers(t *testing.T) {
	pool := newTestPool(t)
	store := NewStore(pool)
	a := newTrainer(t, pool, "เอ")
	b := newTrainer(t, pool, "บี")

	got, err := store.ListTrainers(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []Trainer{{ID: a, Name: "เอ"}, {ID: b, Name: "บี"}}, got)
}

// TestStoreCreateAndList: เพิ่มเวลาเข้างานและเวลาสอนแล้วเรียกดูได้ เรียงตามวันและเวลา (FR-ROST-01, ASM-04).
func TestStoreCreateAndList(t *testing.T) {
	pool := newTestPool(t)
	store := NewStore(pool)
	ctx := context.Background()
	trainer := newTrainer(t, pool, "เอ")

	tuesday := mustCreate(t, store, trainer, KindWork, slot(2, "13:00", "17:00"))
	work := mustCreate(t, store, trainer, KindWork, slot(1, "09:00", "12:00"))
	teach := mustCreate(t, store, trainer, KindTeach, slot(1, "10:00", "11:00"))
	assert.Equal(t, Entry{ID: work.ID, TrainerID: trainer, Kind: KindWork, Slot: slot(1, "09:00", "12:00")}, work)

	got, err := store.ListEntries(ctx, trainer)
	require.NoError(t, err)
	assert.Equal(t, []Entry{work, teach, tuesday}, got)

	raw, err := json.Marshal(teach)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id":`+itoa(teach.ID)+`,"trainer_id":`+itoa(trainer)+
		`,"kind":"teach","weekday":1,"start_time":"10:00","end_time":"11:00","cancelled":false}`, string(raw))
}

// TestStoreCreateRejectsOverlap: ประเภทเดียวกันทับซ้อนถูกปฏิเสธ ต่อกันพอดีได้ และต่างประเภทไม่นับว่าทับ (ASM-02, ASM-10).
func TestStoreCreateRejectsOverlap(t *testing.T) {
	pool := newTestPool(t)
	store := NewStore(pool)
	ctx := context.Background()
	trainer := newTrainer(t, pool, "เอ")
	mustCreate(t, store, trainer, KindWork, slot(1, "09:00", "12:00"))

	_, err := store.Create(ctx, trainer, KindWork, slot(1, "11:00", "13:00"))
	assert.ErrorIs(t, err, ErrOverlap)
	assert.Equal(t, 1, countEntries(t, pool, trainer))

	mustCreate(t, store, trainer, KindWork, slot(1, "12:00", "15:00")) // ต่อกันพอดี (ASM-10)
	mustCreate(t, store, trainer, KindTeach, slot(1, "10:00", "11:00"))
	mustCreate(t, store, trainer, KindTeach, slot(1, "11:00", "12:00")) // เวลาสอนต่อกันพอดีได้ (ASM-10)

	_, err = store.Create(ctx, trainer, KindTeach, slot(1, "10:30", "11:30"))
	assert.ErrorIs(t, err, ErrOverlap)
	assert.Equal(t, 4, countEntries(t, pool, trainer))

	other := newTrainer(t, pool, "บี") // เทรนเนอร์คนอื่นเวลาเดียวกันไม่เกี่ยว
	mustCreate(t, store, other, KindWork, slot(1, "09:00", "12:00"))
}

// TestStoreCreateTeachOutsideWork: เวลาสอนต้องอยู่ในเวลาเข้างาน (รวมช่วงที่ต่อกัน) และเวลาเข้างานที่ยกเลิกไม่นับ (ASM-01, ASM-04, ASM-10).
func TestStoreCreateTeachOutsideWork(t *testing.T) {
	pool := newTestPool(t)
	store := NewStore(pool)
	ctx := context.Background()
	trainer := newTrainer(t, pool, "เอ")

	_, err := store.Create(ctx, trainer, KindTeach, slot(1, "10:00", "11:00"))
	assert.ErrorIs(t, err, ErrTeachOutsideWork, "ยังไม่มีเวลาเข้างาน")

	mustCreate(t, store, trainer, KindWork, slot(1, "09:00", "12:00"))
	mustCreate(t, store, trainer, KindWork, slot(1, "12:00", "15:00"))
	mustCreate(t, store, trainer, KindTeach, slot(1, "11:30", "12:30")) // คร่อมสองช่วงที่ต่อกัน

	_, err = store.Create(ctx, trainer, KindTeach, slot(1, "14:30", "15:30"))
	assert.ErrorIs(t, err, ErrTeachOutsideWork)

	// เวลาเข้างานที่ยกเลิกแล้วไม่นับ (ใส่ตรงในฐานข้อมูล เพราะการยกเลิกผ่าน store เป็นงานของ T-04)
	_, err = pool.Exec(ctx, `INSERT INTO roster_entries (trainer_id, kind, weekday, start_time, end_time, cancelled)
		VALUES ($1, 'work', 3, '09:00', '12:00', true)`, trainer)
	require.NoError(t, err)
	_, err = store.Create(ctx, trainer, KindTeach, slot(3, "10:00", "11:00"))
	assert.ErrorIs(t, err, ErrTeachOutsideWork)
	assert.Equal(t, 4, countEntries(t, pool, trainer))
}

// TestStoreCreateInvalidAndMissingTrainer: รูปแบบผิดได้ ErrInvalidEntry และไม่พบเทรนเนอร์ได้ ErrTrainerNotFound (ASM-02, ASM-03, ASM-09).
func TestStoreCreateInvalidAndMissingTrainer(t *testing.T) {
	pool := newTestPool(t)
	store := NewStore(pool)
	ctx := context.Background()
	trainer := newTrainer(t, pool, "เอ")

	_, err := store.Create(ctx, trainer, KindWork, slot(1, "12:00", "09:00"))
	assert.ErrorIs(t, err, ErrInvalidEntry)
	_, err = store.Create(ctx, trainer, Kind("lunch"), slot(1, "09:00", "12:00"))
	assert.ErrorIs(t, err, ErrInvalidEntry)
	_, err = store.Create(ctx, trainer, KindWork, Slot{Weekday: 9, Start: tod("09:00"), End: tod("12:00")})
	assert.ErrorIs(t, err, ErrInvalidEntry)
	assert.Zero(t, countEntries(t, pool, trainer))

	_, err = store.Create(ctx, trainer+1000, KindWork, slot(1, "09:00", "12:00"))
	assert.ErrorIs(t, err, ErrTrainerNotFound)
	_, err = store.ListEntries(ctx, trainer+1000)
	assert.ErrorIs(t, err, ErrTrainerNotFound)
}

// itoa แปลง int64 เป็นข้อความสำหรับประกอบ JSON ที่คาดหวัง.
func itoa(n int64) string {
	b, _ := json.Marshal(n)
	return string(b)
}
