package inventory

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"gymapp/backend/internal/testutil"
)

// TestAC_INV_01 verifies that a newly created inventory item appears in the catalog with quantity zero (FR-INV-01, CON-DB-01).
func TestAC_INV_01(t *testing.T) {
	pool := testutil.ConnectPool(t)
	router := newCatalogTestRouterWithStore(pool)
	name := uniqueCatalogTestName(t, "AC-INV-01")

	response := performCatalogRequest(t, router, http.MethodPost, "/api/inventory", map[string]any{
		"name":      name,
		"item_type": "equipment",
	})
	require.Equal(t, http.StatusCreated, response.Code)

	var created inventoryItemResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &created))
	require.Equal(t, name, created.Name)
	require.Zero(t, created.QuantityOnHand)
	registerCatalogItemCleanup(t, pool, created.ID)

	response = performCatalogRequest(t, router, http.MethodGet, "/api/inventory", nil)
	require.Equal(t, http.StatusOK, response.Code)
	var items []inventoryItemResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &items))
	require.Contains(t, items, created)
}

// TestAC_INV_02 verifies that an item update is visible when the item is read again (FR-INV-01, CON-DB-01).
func TestAC_INV_02(t *testing.T) {
	pool := testutil.ConnectPool(t)
	router := newCatalogTestRouterWithStore(pool)
	item := createCatalogTestItem(t, pool, router, "AC-INV-02")

	response := performCatalogRequest(t, router, http.MethodPatch, fmt.Sprintf("/api/inventory/%d", item.ID), map[string]any{
		"name": "AC-INV-02 updated item",
	})
	require.Equal(t, http.StatusOK, response.Code)

	response = performCatalogRequest(t, router, http.MethodGet, fmt.Sprintf("/api/inventory/%d", item.ID), nil)
	require.Equal(t, http.StatusOK, response.Code)
	var updated inventoryItemResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &updated))
	require.Equal(t, "AC-INV-02 updated item", updated.Name)
	require.Equal(t, item.ItemType, updated.ItemType)
}

// TestAC_INV_03 verifies that a hidden item is absent from the catalog while its movement history remains (FR-INV-01, CON-DB-01).
func TestAC_INV_03(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	router := newCatalogTestRouterWithStore(pool)
	item := createCatalogTestItem(t, pool, router, "AC-INV-03")

	_, err := NewStore(pool).CreateMovement(ctx, InventoryMovement{
		InventoryItemID: item.ID,
		MovementType:    "receive",
		Quantity:        1,
	})
	require.NoError(t, err)

	response := performCatalogRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/inventory/%d", item.ID), nil)
	require.Equal(t, http.StatusNoContent, response.Code)

	response = performCatalogRequest(t, router, http.MethodGet, "/api/inventory", nil)
	require.Equal(t, http.StatusOK, response.Code)
	var items []inventoryItemResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &items))
	for _, listed := range items {
		require.NotEqual(t, item.ID, listed.ID)
	}

	history, err := NewStore(pool).ListMovements(ctx, item.ID)
	require.NoError(t, err)
	require.Len(t, history, 1)
}

// TestAC_INV_04 verifies that searching a multi-item catalog returns the matching item (FR-INV-01, CON-DB-01).
func TestAC_INV_04(t *testing.T) {
	pool := testutil.ConnectPool(t)
	router := newCatalogTestRouterWithStore(pool)
	target := createCatalogTestItem(t, pool, router, "AC-INV-04 search target")
	createCatalogTestItem(t, pool, router, "AC-INV-04 unrelated")

	response := performCatalogRequest(t, router, http.MethodGet, "/api/inventory?name=SEARCH%20TARGET", nil)
	require.Equal(t, http.StatusOK, response.Code)
	var items []inventoryItemResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &items))
	require.Contains(t, items, target)
	require.Len(t, items, 1)
}

// TestAC_INV_05 verifies that the detail route returns the current quantity of ten (FR-INV-01, CON-DB-01).
func TestAC_INV_05(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	router := newCatalogTestRouterWithStore(pool)
	item := createCatalogTestItem(t, pool, router, "AC-INV-05")

	_, err := pool.Exec(ctx, `UPDATE inventory_items SET quantity_on_hand = 10 WHERE id = $1`, item.ID)
	require.NoError(t, err)

	response := performCatalogRequest(t, router, http.MethodGet, fmt.Sprintf("/api/inventory/%d", item.ID), nil)
	require.Equal(t, http.StatusOK, response.Code)
	var detailed inventoryItemResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &detailed))
	require.Equal(t, int64(10), detailed.QuantityOnHand)
}

// TestAC_INV_06 verifies issuing three consumable units from ten records one issue and leaves seven (FR-INV-02, DOM-INV-01, NFR-DATA-01, CON-DB-01).
func TestAC_INV_06(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	router := newCatalogTestRouterWithStore(pool)
	item := createCatalogTestItem(t, pool, router, "AC-INV-06 consumable")
	_, err := pool.Exec(ctx, `UPDATE inventory_items SET item_type = 'consumable', quantity_on_hand = 10 WHERE id = $1`, item.ID)
	require.NoError(t, err)

	response := performCatalogRequest(t, router, http.MethodPost, fmt.Sprintf("/api/inventory/%d/movements", item.ID), map[string]any{
		"movement_type":      movementTypeIssue,
		"quantity":           3,
		"responsible_person": "T-05 test",
		"purpose":            "Verify transactional issue",
	})
	require.Equal(t, http.StatusCreated, response.Code)
	var result createMovementResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
	require.Equal(t, int64(7), result.QuantityOnHand)
	require.Equal(t, movementTypeIssue, result.Movement.MovementType)
	require.Equal(t, int64(3), result.Movement.Quantity)
	require.Equal(t, "T-05 test", *result.Movement.ResponsiblePerson)
	require.Equal(t, "Verify transactional issue", *result.Movement.Purpose)

	history, err := NewStore(pool).ListMovements(ctx, item.ID)
	require.NoError(t, err)
	require.Len(t, history, 1)
	require.Equal(t, result.Movement.ID, history[0].ID)
	require.Equal(t, movementTypeIssue, history[0].MovementType)
}

// TestAC_INV_09 verifies receiving ten units from five records one receipt and leaves fifteen (FR-INV-02, NFR-DATA-01, CON-DB-01).
func TestAC_INV_09(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	router := newCatalogTestRouterWithStore(pool)
	item := createCatalogTestItem(t, pool, router, "T-05 receive")
	_, err := pool.Exec(ctx, `UPDATE inventory_items SET quantity_on_hand = 5 WHERE id = $1`, item.ID)
	require.NoError(t, err)

	response := performCatalogRequest(t, router, http.MethodPost, fmt.Sprintf("/api/inventory/%d/movements", item.ID), map[string]any{
		"movement_type": movementTypeReceive,
		"quantity":      10,
	})
	require.Equal(t, http.StatusCreated, response.Code)
	var result createMovementResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
	require.Equal(t, int64(15), result.QuantityOnHand)

	history, err := NewStore(pool).ListMovements(ctx, item.ID)
	require.NoError(t, err)
	require.Len(t, history, 1)
	require.Equal(t, movementTypeReceive, history[0].MovementType)
	require.Equal(t, int64(10), history[0].Quantity)
}

// TestAC_INV_07 verifies borrowing one unit from five records a due date and leaves four (FR-INV-02, DOM-INV-01, CON-DB-01).
func TestAC_INV_07(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	router := newCatalogTestRouterWithStore(pool)
	item := createCatalogTestItem(t, pool, router, "AC-INV-07")
	_, err := pool.Exec(ctx, `UPDATE inventory_items SET quantity_on_hand = 5 WHERE id = $1`, item.ID)
	require.NoError(t, err)

	response := performCatalogRequest(t, router, http.MethodPost, fmt.Sprintf("/api/inventory/%d/movements", item.ID), map[string]any{
		"movement_type":      movementTypeBorrow,
		"quantity":           1,
		"responsible_person": "AC-INV-07 borrower",
		"purpose":            "Verify equipment borrow",
		"due_date":           "2027-01-15",
	})
	require.Equal(t, http.StatusCreated, response.Code)
	var result createMovementResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
	require.Equal(t, int64(4), result.QuantityOnHand)
	require.Equal(t, movementTypeBorrow, result.Movement.MovementType)
	require.Equal(t, int64(1), result.Movement.Quantity)
	require.NotNil(t, result.Movement.ResponsiblePerson)
	require.Equal(t, "AC-INV-07 borrower", *result.Movement.ResponsiblePerson)
	require.NotNil(t, result.Movement.Purpose)
	require.Equal(t, "Verify equipment borrow", *result.Movement.Purpose)
	require.NotNil(t, result.Movement.DueDate)
	require.Equal(t, "2027-01-15", *result.Movement.DueDate)

	history, err := NewStore(pool).ListMovements(ctx, item.ID)
	require.NoError(t, err)
	require.Len(t, history, 1)
	require.Equal(t, result.Movement.ID, history[0].ID)
	require.Equal(t, movementTypeBorrow, history[0].MovementType)
}

// TestAC_INV_08 verifies returning borrowed equipment restores the quantity and records one return (FR-INV-02, CON-DB-01).
func TestAC_INV_08(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	router := newCatalogTestRouterWithStore(pool)
	item := createCatalogTestItem(t, pool, router, "AC-INV-08")
	_, err := pool.Exec(ctx, `UPDATE inventory_items SET quantity_on_hand = 5 WHERE id = $1`, item.ID)
	require.NoError(t, err)

	borrowResponse := performCatalogRequest(t, router, http.MethodPost, fmt.Sprintf("/api/inventory/%d/movements", item.ID), map[string]any{
		"movement_type":      movementTypeBorrow,
		"quantity":           1,
		"responsible_person": "AC-INV-08 borrower",
		"purpose":            "Prepare borrowed equipment",
		"due_date":           "2027-01-15",
	})
	require.Equal(t, http.StatusCreated, borrowResponse.Code)
	var borrowed createMovementResponse
	require.NoError(t, json.Unmarshal(borrowResponse.Body.Bytes(), &borrowed))
	require.Equal(t, int64(4), borrowed.QuantityOnHand)

	response := performCatalogRequest(t, router, http.MethodPost, fmt.Sprintf("/api/inventory/%d/movements", item.ID), map[string]any{
		"movement_type":      movementTypeReturn,
		"quantity":           1,
		"responsible_person": "AC-INV-08 borrower",
		"purpose":            "Record equipment return",
	})
	require.Equal(t, http.StatusCreated, response.Code)
	var result createMovementResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
	require.Equal(t, int64(5), result.QuantityOnHand)
	require.Equal(t, movementTypeReturn, result.Movement.MovementType)
	require.Equal(t, int64(1), result.Movement.Quantity)

	history, err := NewStore(pool).ListMovements(ctx, item.ID)
	require.NoError(t, err)
	require.Len(t, history, 2)
	require.Equal(t, movementTypeBorrow, history[0].MovementType)
	require.Equal(t, movementTypeReturn, history[1].MovementType)
	require.Equal(t, result.Movement.ID, history[1].ID)
	require.Equal(t, borrowed.Movement.ID, history[0].ID)
	require.Nil(t, history[1].DueDate)
}

// TestAC_INV_10 rejects issue and borrow requests without a responsible person and preserves stock and history (DOM-INV-01, CON-DB-01).
func TestAC_INV_10(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	router := newCatalogTestRouterWithStore(pool)
	store := NewStore(pool)
	cases := []struct {
		name         string
		movementType string
		itemType     string
	}{
		{name: "issue", movementType: movementTypeIssue, itemType: "consumable"},
		{name: "borrow", movementType: movementTypeBorrow, itemType: "equipment"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			item := createCatalogTestItem(t, pool, router, "AC-INV-10 "+tc.name)
			_, err := pool.Exec(ctx, `UPDATE inventory_items SET item_type = $2, quantity_on_hand = 10 WHERE id = $1`, item.ID, tc.itemType)
			require.NoError(t, err)

			request := map[string]any{
				"movement_type": tc.movementType,
				"quantity":      1,
				"purpose":       "Verify missing responsible person is rejected",
			}
			if tc.movementType == movementTypeBorrow {
				request["due_date"] = "2027-01-15"
			}
			response := performCatalogRequest(t, router, http.MethodPost, fmt.Sprintf("/api/inventory/%d/movements", item.ID), request)
			require.Equal(t, http.StatusBadRequest, response.Code)

			current, err := store.GetItem(ctx, item.ID)
			require.NoError(t, err)
			require.Equal(t, int64(10), current.QuantityOnHand)
			history, err := store.ListMovements(ctx, item.ID)
			require.NoError(t, err)
			require.Empty(t, history)
		})
	}
}

// TestAC_INV_11 rejects issue and borrow requests without a purpose and preserves stock and history (DOM-INV-01, CON-DB-01).
func TestAC_INV_11(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	router := newCatalogTestRouterWithStore(pool)
	store := NewStore(pool)
	cases := []struct {
		name         string
		movementType string
		itemType     string
	}{
		{name: "issue", movementType: movementTypeIssue, itemType: "consumable"},
		{name: "borrow", movementType: movementTypeBorrow, itemType: "equipment"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			item := createCatalogTestItem(t, pool, router, "AC-INV-11 "+tc.name)
			_, err := pool.Exec(ctx, `UPDATE inventory_items SET item_type = $2, quantity_on_hand = 10 WHERE id = $1`, item.ID, tc.itemType)
			require.NoError(t, err)

			request := map[string]any{
				"movement_type":      tc.movementType,
				"quantity":           1,
				"responsible_person": "AC-INV-11 requester",
			}
			if tc.movementType == movementTypeBorrow {
				request["due_date"] = "2027-01-15"
			}
			response := performCatalogRequest(t, router, http.MethodPost, fmt.Sprintf("/api/inventory/%d/movements", item.ID), request)
			require.Equal(t, http.StatusBadRequest, response.Code)

			current, err := store.GetItem(ctx, item.ID)
			require.NoError(t, err)
			require.Equal(t, int64(10), current.QuantityOnHand)
			history, err := store.ListMovements(ctx, item.ID)
			require.NoError(t, err)
			require.Empty(t, history)
		})
	}
}

// TestAC_INV_12 rejects a borrow request without a due date and preserves the quantity and movement history (DOM-INV-01, CON-DB-01).
func TestAC_INV_12(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	router := newCatalogTestRouterWithStore(pool)
	store := NewStore(pool)
	item := createCatalogTestItem(t, pool, router, "AC-INV-12")
	_, err := pool.Exec(ctx, `UPDATE inventory_items SET quantity_on_hand = 5 WHERE id = $1`, item.ID)
	require.NoError(t, err)

	response := performCatalogRequest(t, router, http.MethodPost, fmt.Sprintf("/api/inventory/%d/movements", item.ID), map[string]any{
		"movement_type":      movementTypeBorrow,
		"quantity":           1,
		"responsible_person": "AC-INV-12 borrower",
		"purpose":            "Verify missing due date is rejected",
	})
	require.Equal(t, http.StatusBadRequest, response.Code)

	current, err := store.GetItem(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, int64(5), current.QuantityOnHand)
	history, err := store.ListMovements(ctx, item.ID)
	require.NoError(t, err)
	require.Empty(t, history)
}

// TestLowStockAlertAPI returns only visible items at or below a configured reorder point (FR-INV-03, DOM-INV-02, CON-DB-01).
func TestLowStockAlertAPI(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	router := newCatalogTestRouterWithStore(pool)
	equal := createCatalogTestItem(t, pool, router, "T-12 alert equal")
	below := createCatalogTestItem(t, pool, router, "T-12 alert below")
	above := createCatalogTestItem(t, pool, router, "T-12 alert above")
	unconfigured := createCatalogTestItem(t, pool, router, "T-12 alert unset")

	for _, item := range []inventoryItemResponse{equal, below, above} {
		_, err := pool.Exec(ctx, `UPDATE inventory_items SET reorder_point = 5 WHERE id = $1`, item.ID)
		require.NoError(t, err)
	}
	for _, item := range []inventoryItemResponse{equal, below, above, unconfigured} {
		_, err := pool.Exec(ctx, `UPDATE inventory_items SET quantity_on_hand = $2 WHERE id = $1`, item.ID, map[int64]int64{
			equal.ID:        5,
			below.ID:        4,
			above.ID:        6,
			unconfigured.ID: 0,
		}[item.ID])
		require.NoError(t, err)
	}

	response := performCatalogRequest(t, router, http.MethodGet, "/api/inventory/alerts", nil)
	require.Equal(t, http.StatusOK, response.Code)
	var alerts []inventoryItemResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &alerts))
	require.Len(t, alerts, 2)
	require.Equal(t, []int64{equal.ID, below.ID}, []int64{alerts[0].ID, alerts[1].ID})
	require.Equal(t, int64(5), alerts[0].QuantityOnHand)
	require.Equal(t, int64(4), alerts[1].QuantityOnHand)
}

// TestAC_INV_13 verifies that issuing stock down to the reorder point alerts for item A but not for item B without a reorder point (FR-INV-03, DOM-INV-02, CON-DB-01).
func TestAC_INV_13(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	router := newCatalogTestRouterWithStore(pool)
	itemA := createCatalogTestItem(t, pool, router, "AC-INV-13 item A")
	itemB := createCatalogTestItem(t, pool, router, "AC-INV-13 item B")
	_, err := pool.Exec(ctx, `
		UPDATE inventory_items
		SET reorder_point = 5, quantity_on_hand = 6
		WHERE id = $1`, itemA.ID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		UPDATE inventory_items
		SET reorder_point = NULL, quantity_on_hand = 5
		WHERE id = $1`, itemB.ID)
	require.NoError(t, err)

	movementResponse := performCatalogRequest(t, router, http.MethodPost, fmt.Sprintf("/api/inventory/%d/movements", itemA.ID), map[string]any{
		"movement_type":      movementTypeIssue,
		"quantity":           1,
		"responsible_person": "AC-INV-13 test",
		"purpose":            "Reach the reorder point",
	})
	require.Equal(t, http.StatusCreated, movementResponse.Code)
	var movement createMovementResponse
	require.NoError(t, json.Unmarshal(movementResponse.Body.Bytes(), &movement))
	require.Equal(t, int64(5), movement.QuantityOnHand)

	alertResponse := performCatalogRequest(t, router, http.MethodGet, "/api/inventory/alerts", nil)
	require.Equal(t, http.StatusOK, alertResponse.Code)
	var alerts []inventoryItemResponse
	require.NoError(t, json.Unmarshal(alertResponse.Body.Bytes(), &alerts))

	alertIDs := make([]int64, 0, len(alerts))
	for _, alert := range alerts {
		alertIDs = append(alertIDs, alert.ID)
	}
	require.Contains(t, alertIDs, itemA.ID)
	require.NotContains(t, alertIDs, itemB.ID)

	var alertedItem inventoryItemResponse
	for _, alert := range alerts {
		if alert.ID == itemA.ID {
			alertedItem = alert
			break
		}
	}
	require.Equal(t, int64(5), alertedItem.QuantityOnHand)
	require.NotNil(t, alertedItem.ReorderPoint)
	require.Equal(t, int64(5), *alertedItem.ReorderPoint)
}

// TestMovementDatabaseFailureRollsBack verifies a movement insert failure returns an error and rolls back the quantity update (CON-DB-01, NFR-DATA-01, ASM-07).
func TestMovementDatabaseFailureRollsBack(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	router := newCatalogTestRouterWithStore(pool)
	store := NewStore(pool)
	item := createCatalogTestItem(t, pool, router, "T-10 database failure")
	_, err := pool.Exec(ctx, `UPDATE inventory_items SET quantity_on_hand = 10 WHERE id = $1`, item.ID)
	require.NoError(t, err)

	functionName := strings.ReplaceAll(uniqueCatalogTestName(t, "t10_fail_movement"), " ", "_")
	triggerName := functionName + "_trigger"
	_, err = pool.Exec(ctx, fmt.Sprintf(`
		CREATE FUNCTION %s() RETURNS trigger
		LANGUAGE plpgsql
		AS $$
		BEGIN
			RAISE EXCEPTION 'simulated inventory movement insert failure';
		END;
		$$`, functionName))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, cleanupErr := pool.Exec(ctx, fmt.Sprintf(`DROP FUNCTION IF EXISTS %s()`, functionName))
		require.NoError(t, cleanupErr)
	})

	_, err = pool.Exec(ctx, fmt.Sprintf(`
		CREATE TRIGGER %s
		BEFORE INSERT ON inventory_movements
		FOR EACH ROW EXECUTE FUNCTION %s()`, triggerName, functionName))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, cleanupErr := pool.Exec(ctx, fmt.Sprintf(`DROP TRIGGER IF EXISTS %s ON inventory_movements`, triggerName))
		require.NoError(t, cleanupErr)
	})

	response := performCatalogRequest(t, router, http.MethodPost, fmt.Sprintf("/api/inventory/%d/movements", item.ID), map[string]any{
		"movement_type":      movementTypeIssue,
		"quantity":           3,
		"responsible_person": "T-10 test",
		"purpose":            "Verify database failure rollback",
	})
	require.Equal(t, http.StatusInternalServerError, response.Code)
	var errorResponse map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &errorResponse))
	require.Equal(t, "internal_error", errorResponse["error"])

	current, err := store.GetItem(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, int64(10), current.QuantityOnHand)
	history, err := store.ListMovements(ctx, item.ID)
	require.NoError(t, err)
	require.Empty(t, history)
}

// TestBorrowReturnMovementFlow verifies borrowing reduces stock and return restores it with both movements recorded (FR-INV-02, DOM-INV-01, NFR-DATA-01, CON-DB-01).
func TestBorrowReturnMovementFlow(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	router := newCatalogTestRouterWithStore(pool)
	item := createCatalogTestItem(t, pool, router, "T-07 borrow-return")
	_, err := pool.Exec(ctx, `UPDATE inventory_items SET quantity_on_hand = 5 WHERE id = $1`, item.ID)
	require.NoError(t, err)

	borrowResponse := performCatalogRequest(t, router, http.MethodPost, fmt.Sprintf("/api/inventory/%d/movements", item.ID), map[string]any{
		"movement_type":      movementTypeBorrow,
		"quantity":           1,
		"responsible_person": "T-07 test",
		"purpose":            "Verify borrow and return",
		"due_date":           "2027-01-15",
	})
	require.Equal(t, http.StatusCreated, borrowResponse.Code)
	var borrowed createMovementResponse
	require.NoError(t, json.Unmarshal(borrowResponse.Body.Bytes(), &borrowed))
	require.Equal(t, int64(4), borrowed.QuantityOnHand)
	require.Equal(t, movementTypeBorrow, borrowed.Movement.MovementType)
	require.NotNil(t, borrowed.Movement.DueDate)
	require.Equal(t, "2027-01-15", *borrowed.Movement.DueDate)

	returnResponse := performCatalogRequest(t, router, http.MethodPost, fmt.Sprintf("/api/inventory/%d/movements", item.ID), map[string]any{
		"movement_type":      movementTypeReturn,
		"quantity":           1,
		"responsible_person": "T-07 test",
		"purpose":            "Verify returned equipment",
	})
	require.Equal(t, http.StatusCreated, returnResponse.Code)
	var returned createMovementResponse
	require.NoError(t, json.Unmarshal(returnResponse.Body.Bytes(), &returned))
	require.Equal(t, int64(5), returned.QuantityOnHand)
	require.Equal(t, movementTypeReturn, returned.Movement.MovementType)

	history, err := NewStore(pool).ListMovements(ctx, item.ID)
	require.NoError(t, err)
	require.Len(t, history, 2)
	require.Equal(t, movementTypeBorrow, history[0].MovementType)
	require.Equal(t, movementTypeReturn, history[1].MovementType)
	require.NotNil(t, history[0].DueDate)
	require.Equal(t, "2027-01-15", history[0].DueDate.Format("2006-01-02"))
	require.Nil(t, history[1].DueDate)
}

// TestStockMovementInsufficientStockRollsBack verifies a rejected issue changes neither quantity nor movement history (FR-INV-02, NFR-DATA-01, CON-DB-01).
func TestStockMovementInsufficientStockRollsBack(t *testing.T) {
	ctx := context.Background()
	pool := testutil.ConnectPool(t)
	router := newCatalogTestRouterWithStore(pool)
	item := createCatalogTestItem(t, pool, router, "T-05 insufficient")
	_, err := pool.Exec(ctx, `UPDATE inventory_items SET quantity_on_hand = 2 WHERE id = $1`, item.ID)
	require.NoError(t, err)

	response := performCatalogRequest(t, router, http.MethodPost, fmt.Sprintf("/api/inventory/%d/movements", item.ID), map[string]any{
		"movement_type":      movementTypeIssue,
		"quantity":           3,
		"responsible_person": "T-05 test",
		"purpose":            "Verify rejected issue",
	})
	require.Equal(t, http.StatusConflict, response.Code)

	current, err := NewStore(pool).GetItem(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), current.QuantityOnHand)
	history, err := NewStore(pool).ListMovements(ctx, item.ID)
	require.NoError(t, err)
	require.Empty(t, history)
}

// performCatalogRequest sends one JSON HTTP request to a catalog route for AC verification (FR-INV-01, CON-TECH-01).
func performCatalogRequest(t *testing.T, router http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var requestBody bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&requestBody).Encode(body))
	}

	request := httptest.NewRequest(method, path, &requestBody)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

// createCatalogTestItem creates a uniquely named item over the catalog API and removes it after AC verification (FR-INV-01, CON-DB-01).
func createCatalogTestItem(t *testing.T, pool *pgxpool.Pool, router http.Handler, prefix string) inventoryItemResponse {
	t.Helper()
	name := uniqueCatalogTestName(t, prefix)
	response := performCatalogRequest(t, router, http.MethodPost, "/api/inventory", map[string]any{
		"name":      name,
		"item_type": "equipment",
	})
	require.Equal(t, http.StatusCreated, response.Code)

	var item inventoryItemResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &item))
	registerCatalogItemCleanup(t, pool, item.ID)
	return item
}

// newCatalogTestRouterWithStore registers the catalog routes against the PostgreSQL-backed store (FR-INV-01, CON-TECH-01, CON-DB-01).
func newCatalogTestRouterWithStore(pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewHandler(NewStore(pool)).Register(router)
	return router
}

// uniqueCatalogTestName isolates records created for a single AC test (FR-INV-01, CON-DB-01).
func uniqueCatalogTestName(t *testing.T, prefix string) string {
	t.Helper()
	random := make([]byte, 8)
	_, err := rand.Read(random)
	require.NoError(t, err)
	return prefix + " " + hex.EncodeToString(random)
}

// registerCatalogItemCleanup removes only the item and movements created by a catalog AC test (FR-INV-01, CON-DB-01).
func registerCatalogItemCleanup(t *testing.T, pool *pgxpool.Pool, itemID int64) {
	t.Helper()
	t.Cleanup(func() {
		ctx := context.Background()
		_, err := pool.Exec(ctx, `DELETE FROM inventory_movements WHERE inventory_item_id = $1`, itemID)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `DELETE FROM inventory_items WHERE id = $1`, itemID)
		require.NoError(t, err)
	})
}
