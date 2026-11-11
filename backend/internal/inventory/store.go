package inventory

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrItemNotFound indicates that an item is unavailable in the visible catalog (FR-INV-01).
var ErrItemNotFound = errors.New("inventory: item not found")

// Store reads and writes inventory items and movements in PostgreSQL (FR-INV-01, FR-INV-02, CON-DB-01).
type Store struct {
	pool *pgxpool.Pool
}

// NewStore creates an inventory store using the caller-owned PostgreSQL pool (FR-INV-01, FR-INV-02, CON-DB-01).
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// CreateItem inserts an inventory item with an initial quantity of zero (FR-INV-01, CON-DB-01).
func (s *Store) CreateItem(ctx context.Context, item InventoryItem) (InventoryItem, error) {
	created, err := scanInventoryItem(s.pool.QueryRow(ctx, `
		INSERT INTO inventory_items (name, item_type, reorder_point)
		VALUES ($1, $2, $3)
		RETURNING id, name, item_type, quantity_on_hand, reorder_point, deleted_at`,
		item.Name, item.ItemType, item.ReorderPoint))
	if err != nil {
		return InventoryItem{}, fmt.Errorf("inventory: create item: %w", err)
	}
	return created, nil
}

// GetItem retrieves an item that is visible in the catalog (FR-INV-01, CON-DB-01).
func (s *Store) GetItem(ctx context.Context, id int64) (InventoryItem, error) {
	item, err := scanInventoryItem(s.pool.QueryRow(ctx, `
		SELECT id, name, item_type, quantity_on_hand, reorder_point, deleted_at
		FROM inventory_items
		WHERE id = $1 AND deleted_at IS NULL`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return InventoryItem{}, ErrItemNotFound
		}
		return InventoryItem{}, fmt.Errorf("inventory: get item: %w", err)
	}
	return item, nil
}

// ListItems searches visible catalog items by a case-insensitive partial name (FR-INV-01, CON-DB-01).
func (s *Store) ListItems(ctx context.Context, name string) ([]InventoryItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, item_type, quantity_on_hand, reorder_point, deleted_at
		FROM inventory_items
		WHERE deleted_at IS NULL AND name ILIKE '%' || $1 || '%'
		ORDER BY id`, name)
	if err != nil {
		return nil, fmt.Errorf("inventory: list items: %w", err)
	}
	defer rows.Close()

	items := make([]InventoryItem, 0)
	for rows.Next() {
		item, err := scanInventoryItem(rows)
		if err != nil {
			return nil, fmt.Errorf("inventory: scan item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("inventory: list items: %w", err)
	}
	return items, nil
}

// ListLowStockItems returns visible inventory items at or below their reorder point (FR-INV-03, DOM-INV-02, CON-DB-01).
func (s *Store) ListLowStockItems(ctx context.Context) ([]InventoryItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, item_type, quantity_on_hand, reorder_point, deleted_at
		FROM inventory_items
		WHERE deleted_at IS NULL
		  AND reorder_point IS NOT NULL
		  AND quantity_on_hand <= reorder_point
		ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("inventory: list low stock items: %w", err)
	}
	defer rows.Close()

	items := make([]InventoryItem, 0)
	for rows.Next() {
		item, err := scanInventoryItem(rows)
		if err != nil {
			return nil, fmt.Errorf("inventory: scan low stock item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("inventory: iterate low stock items: %w", err)
	}
	return items, nil
}

// UpdateItem replaces the catalog fields of a visible item (FR-INV-01, DOM-INV-02, CON-DB-01).
func (s *Store) UpdateItem(ctx context.Context, item InventoryItem) (InventoryItem, error) {
	updated, err := scanInventoryItem(s.pool.QueryRow(ctx, `
		UPDATE inventory_items
		SET name = $2, item_type = $3, reorder_point = $4
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, name, item_type, quantity_on_hand, reorder_point, deleted_at`,
		item.ID, item.Name, item.ItemType, item.ReorderPoint))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return InventoryItem{}, ErrItemNotFound
		}
		return InventoryItem{}, fmt.Errorf("inventory: update item: %w", err)
	}
	return updated, nil
}

// SoftDeleteItem hides an item from the catalog without deleting its movement history (FR-INV-01, FR-INV-02, CON-DB-01).
func (s *Store) SoftDeleteItem(ctx context.Context, id int64) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE inventory_items
		SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("inventory: hide item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrItemNotFound
	}
	return nil
}

// CreateMovement persists an inventory movement without changing the item's quantity (FR-INV-02, CON-DB-01).
func (s *Store) CreateMovement(ctx context.Context, movement InventoryMovement) (InventoryMovement, error) {
	created, err := scanInventoryMovement(s.pool.QueryRow(ctx, `
		INSERT INTO inventory_movements
			(inventory_item_id, movement_type, quantity, responsible_person, purpose, due_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, inventory_item_id, movement_type, quantity, responsible_person, purpose, due_date, created_at`,
		movement.InventoryItemID, movement.MovementType, movement.Quantity,
		movement.ResponsiblePerson, movement.Purpose, movement.DueDate))
	if err != nil {
		return InventoryMovement{}, fmt.Errorf("inventory: create movement: %w", err)
	}
	return created, nil
}

// ApplyStockMovement ปรับยอดและเพิ่มประวัติ movement ภายใน PostgreSQL transaction เดียว (FR-INV-02, NFR-DATA-01, CON-DB-01).
func (s *Store) ApplyStockMovement(ctx context.Context, itemID int64, movement InventoryMovement) (created InventoryMovement, quantityOnHand int64, err error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return InventoryMovement{}, 0, fmt.Errorf("inventory: begin movement transaction: %w", err)
	}
	finished := false
	defer func() {
		if finished {
			return
		}
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			err = errors.Join(err, fmt.Errorf("inventory: rollback movement transaction: %w", rollbackErr))
		}
	}()

	var updateQuery string
	switch movement.MovementType {
	case movementTypeIssue, movementTypeBorrow:
		updateQuery = `
			UPDATE inventory_items
			SET quantity_on_hand = quantity_on_hand - $2
			WHERE id = $1 AND deleted_at IS NULL AND quantity_on_hand >= $2`
	case movementTypeReceive, movementTypeReturn:
		updateQuery = `
			UPDATE inventory_items
			SET quantity_on_hand = quantity_on_hand + $2
			WHERE id = $1 AND deleted_at IS NULL`
	default:
		return InventoryMovement{}, 0, ErrInvalidMovement
	}

	tag, err := tx.Exec(ctx, updateQuery, itemID, movement.Quantity)
	if err != nil {
		return InventoryMovement{}, 0, fmt.Errorf("inventory: update stock quantity: %w", err)
	}
	if tag.RowsAffected() == 0 {
		var exists bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM inventory_items WHERE id = $1 AND deleted_at IS NULL
			)`, itemID).Scan(&exists); err != nil {
			return InventoryMovement{}, 0, fmt.Errorf("inventory: check item for movement: %w", err)
		}
		if !exists {
			return InventoryMovement{}, 0, ErrItemNotFound
		}
		return InventoryMovement{}, 0, ErrInsufficientStock
	}

	created, err = scanInventoryMovement(tx.QueryRow(ctx, `
		INSERT INTO inventory_movements
			(inventory_item_id, movement_type, quantity, responsible_person, purpose, due_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, inventory_item_id, movement_type, quantity, responsible_person, purpose, due_date, created_at`,
		itemID, movement.MovementType, movement.Quantity,
		movement.ResponsiblePerson, movement.Purpose, movement.DueDate))
	if err != nil {
		return InventoryMovement{}, 0, fmt.Errorf("inventory: insert movement: %w", err)
	}

	if err := tx.QueryRow(ctx, `SELECT quantity_on_hand FROM inventory_items WHERE id = $1`, itemID).Scan(&quantityOnHand); err != nil {
		return InventoryMovement{}, 0, fmt.Errorf("inventory: read updated quantity: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return InventoryMovement{}, 0, fmt.Errorf("inventory: commit movement transaction: %w", err)
	}
	finished = true
	return created, quantityOnHand, nil
}

// ListMovements retrieves the complete movement history for an item (FR-INV-02, CON-DB-01).
func (s *Store) ListMovements(ctx context.Context, itemID int64) ([]InventoryMovement, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, inventory_item_id, movement_type, quantity, responsible_person, purpose, due_date, created_at
		FROM inventory_movements
		WHERE inventory_item_id = $1
		ORDER BY id`, itemID)
	if err != nil {
		return nil, fmt.Errorf("inventory: list movements: %w", err)
	}
	defer rows.Close()

	movements := make([]InventoryMovement, 0)
	for rows.Next() {
		movement, err := scanInventoryMovement(rows)
		if err != nil {
			return nil, fmt.Errorf("inventory: scan movement: %w", err)
		}
		movements = append(movements, movement)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("inventory: list movements: %w", err)
	}
	return movements, nil
}

// scanInventoryItem reads an inventory item row from PostgreSQL (FR-INV-01, CON-DB-01).
func scanInventoryItem(row pgx.Row) (InventoryItem, error) {
	var item InventoryItem
	err := row.Scan(
		&item.ID, &item.Name, &item.ItemType, &item.QuantityOnHand,
		&item.ReorderPoint, &item.DeletedAt,
	)
	return item, err
}

// scanInventoryMovement reads a movement history row from PostgreSQL (FR-INV-02, CON-DB-01).
func scanInventoryMovement(row pgx.Row) (InventoryMovement, error) {
	var movement InventoryMovement
	err := row.Scan(
		&movement.ID, &movement.InventoryItemID, &movement.MovementType, &movement.Quantity,
		&movement.ResponsiblePerson, &movement.Purpose, &movement.DueDate, &movement.CreatedAt,
	)
	return movement, err
}
