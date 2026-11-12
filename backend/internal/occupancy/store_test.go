package occupancy

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"gymapp/backend/internal/testutil"
)

// testNoon คือเวลาคงที่กลางวัน (Asia/Bangkok) ที่ใช้ใน test ของ store เพื่อไม่ให้ผลขึ้นกับเวลาจริงหรือเที่ยงคืน (ASM-03).
var testNoon = time.Date(2026, 11, 12, 12, 0, 0, 0, bangkok)

// resetGateSignals ลบสัญญาณทั้งหมดในฐานข้อมูลทดสอบ เพราะจำนวนปัจจุบันเป็นค่ารวมของทั้งตาราง
// test ใน package นี้จึงห้ามใช้ t.Parallel (ใช้กับฐานข้อมูลทดสอบเท่านั้น).
func resetGateSignals(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	_, err := pool.Exec(ctx, `DELETE FROM gate_signals`)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM gate_signals`) })
}

// recordN บันทึกสัญญาณทิศทาง d จำนวน n ครั้งที่เวลา at.
func recordN(t *testing.T, store *Store, d Direction, n int, at time.Time) {
	t.Helper()
	for i := 0; i < n; i++ {
		_, err := store.RecordSignal(context.Background(), d, at)
		require.NoError(t, err)
	}
}

// TestAC_BI_02_Store: Given จำนวนปัจจุบัน 10 When รับสัญญาณผ่านเข้า 1 ครั้ง Then จำนวนเป็น 11
// และเมื่อสัญญาณเข้ามาพร้อมกันหลายตัว จำนวนต้องเพิ่มเท่ากับจำนวนสัญญาณ (DOM-CAP-01, ASM-05).
func TestAC_BI_02_Store(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	resetGateSignals(t, pool)
	store := NewStore(pool)

	// Given
	recordN(t, store, DirectionIn, 10, testNoon)
	// When
	got, err := store.RecordSignal(ctx, DirectionIn, testNoon)
	// Then
	require.NoError(t, err)
	require.Equal(t, 11, got)
	count, err := store.CurrentCount(ctx, testNoon)
	require.NoError(t, err)
	require.Equal(t, 11, count)

	t.Run("concurrent signals", func(t *testing.T) {
		const workers = 20
		var wg sync.WaitGroup
		errs := make(chan error, workers)
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := store.RecordSignal(ctx, DirectionIn, testNoon)
				errs <- err
			}()
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			require.NoError(t, err)
		}
		count, err := store.CurrentCount(ctx, testNoon)
		require.NoError(t, err)
		require.Equal(t, 11+workers, count)
	})
}

// TestAC_BI_05_Store: Given จำนวนปัจจุบัน 0 When รับสัญญาณผ่านออก 1 ครั้ง Then จำนวนยังเป็น 0 และสัญญาณนั้นถูกบันทึกไว้ (ASM-04).
func TestAC_BI_05_Store(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	resetGateSignals(t, pool)
	store := NewStore(pool)

	// Given
	count, err := store.CurrentCount(ctx, testNoon)
	require.NoError(t, err)
	require.Equal(t, 0, count)
	// When
	got, err := store.RecordSignal(ctx, DirectionOut, testNoon)
	// Then
	require.NoError(t, err)
	require.Equal(t, 0, got)
	var rows int
	var direction string
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*), max(direction) FROM gate_signals`).Scan(&rows, &direction))
	require.Equal(t, 1, rows)
	require.Equal(t, "out", direction)
}

// TestAC_BI_07_Store: Given จำนวนปัจจุบัน 10 When backend รีสตาร์ต (สร้าง pool และ store ใหม่) ภายในวันเดียวกัน Then จำนวนยังเป็น 10 (ASM-02).
func TestAC_BI_07_Store(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	resetGateSignals(t, pool)

	// Given
	recordN(t, NewStore(pool), DirectionIn, 10, testNoon)
	// When: pool ใหม่ + store ใหม่ = ไม่มีสถานะในหน่วยความจำจาก process เดิม
	restarted, err := pgxpool.New(ctx, testutil.TestDSN())
	require.NoError(t, err)
	defer restarted.Close()
	count, err := NewStore(restarted).CurrentCount(ctx, testNoon.Add(time.Hour))
	// Then
	require.NoError(t, err)
	require.Equal(t, 10, count)
}

// TestStoreRejectsInvalidDirection ยืนยันว่าทิศทางอื่นไม่ถูกบันทึก (IF-GATE-01) — ไม่ต้องต่อฐานข้อมูล.
func TestStoreRejectsInvalidDirection(t *testing.T) {
	_, err := (&Store{}).RecordSignal(context.Background(), Direction("sideways"), testNoon)
	require.ErrorIs(t, err, ErrInvalidDirection)
}
