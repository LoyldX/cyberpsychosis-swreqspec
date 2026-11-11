package inventory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"gymapp/backend/internal/testutil"
)

// TestAC_INV_14 verifies only one of two concurrent issues of three succeeds from a stock of five (NFR-DATA-01, CON-DB-01).
func TestAC_INV_14(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	store := NewStore(pool)

	random := make([]byte, 8)
	_, err := rand.Read(random)
	require.NoError(t, err)
	item, err := store.CreateItem(ctx, InventoryItem{
		Name:     "AC-INV-14 " + hex.EncodeToString(random),
		ItemType: "consumable",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := pool.Exec(ctx, `DELETE FROM inventory_movements WHERE inventory_item_id = $1`, item.ID)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `DELETE FROM inventory_items WHERE id = $1`, item.ID)
		require.NoError(t, err)
	})
	_, err = pool.Exec(ctx, `UPDATE inventory_items SET quantity_on_hand = 5 WHERE id = $1`, item.ID)
	require.NoError(t, err)

	type outcome struct {
		result MovementResult
		err    error
	}
	start := make(chan struct{})
	ready := make(chan struct{}, 2)
	outcomes := make(chan outcome, 2)
	var workers sync.WaitGroup

	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			ready <- struct{}{}
			<-start
			result, err := NewService(store).CreateMovement(ctx, item.ID, MovementCommand{
				Type:              movementTypeIssue,
				Quantity:          3,
				ResponsiblePerson: "AC-INV-14 test",
				Purpose:           "Verify concurrent stock consistency",
			})
			outcomes <- outcome{result: result, err: err}
		}()
	}

	<-ready
	<-ready
	close(start)
	workers.Wait()
	close(outcomes)

	successes := 0
	rejections := 0
	for result := range outcomes {
		switch {
		case result.err == nil:
			successes++
			require.Equal(t, int64(2), result.result.QuantityOnHand)
			require.Equal(t, movementTypeIssue, result.result.Movement.MovementType)
			require.Equal(t, int64(3), result.result.Movement.Quantity)
		case errors.Is(result.err, ErrInsufficientStock):
			rejections++
		default:
			require.NoError(t, result.err, fmt.Sprintf("unexpected movement error: %v", result.err))
		}
	}

	require.Equal(t, 1, successes)
	require.Equal(t, 1, rejections)

	current, err := store.GetItem(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), current.QuantityOnHand)
	history, err := store.ListMovements(ctx, item.ID)
	require.NoError(t, err)
	require.Len(t, history, 1)
	require.Equal(t, movementTypeIssue, history[0].MovementType)
	require.Equal(t, int64(3), history[0].Quantity)
}
