package inventory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gymapp/backend/internal/testutil"
)

func TestStoreItemAndMovementPersistenceAndSoftDelete(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	store := NewStore(pool)

	reorderPoint := int64(5)
	item, err := store.CreateItem(ctx, InventoryItem{
		Name:           "T-02 test item",
		ItemType:       "consumable",
		QuantityOnHand: 99,
		ReorderPoint:   &reorderPoint,
	})
	require.NoError(t, err)
	require.Zero(t, item.QuantityOnHand, "new catalog items start at zero")
	require.Equal(t, &reorderPoint, item.ReorderPoint)

	t.Cleanup(func() {
		_, cleanupErr := pool.Exec(ctx, `DELETE FROM inventory_movements WHERE inventory_item_id = $1`, item.ID)
		require.NoError(t, cleanupErr)
		_, cleanupErr = pool.Exec(ctx, `DELETE FROM inventory_items WHERE id = $1`, item.ID)
		require.NoError(t, cleanupErr)
	})

	responsible := "Inventory test"
	purpose := "Store persistence test"
	movement, err := store.CreateMovement(ctx, InventoryMovement{
		InventoryItemID:   item.ID,
		MovementType:      "receive",
		Quantity:          10,
		ResponsiblePerson: &responsible,
		Purpose:           &purpose,
	})
	require.NoError(t, err)
	require.Equal(t, item.ID, movement.InventoryItemID)

	history, err := store.ListMovements(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, []InventoryMovement{movement}, history)

	require.NoError(t, store.SoftDeleteItem(ctx, item.ID))
	_, err = store.GetItem(ctx, item.ID)
	require.ErrorIs(t, err, ErrItemNotFound)

	catalog, err := store.ListItems(ctx, "T-02 test item")
	require.NoError(t, err)
	require.Empty(t, catalog, "soft-deleted items must not appear in the catalog")

	history, err = store.ListMovements(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, []InventoryMovement{movement}, history, "soft delete must preserve movement history")
}

func TestStoreSearchIsCaseInsensitiveAndPartial(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	store := NewStore(pool)

	item, err := store.CreateItem(ctx, InventoryItem{
		Name:     "Portable Inventory Search",
		ItemType: "equipment",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, cleanupErr := pool.Exec(ctx, `DELETE FROM inventory_items WHERE id = $1`, item.ID)
		require.NoError(t, cleanupErr)
	})

	results, err := store.ListItems(ctx, "inventory se")
	require.NoError(t, err)
	require.Contains(t, results, item)
}
