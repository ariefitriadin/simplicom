// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: warehouse.sql

package pgrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const cancelExpiredReservations = `-- name: CancelExpiredReservations :many
WITH expired_reservations AS (
    DELETE FROM stock_reservations
    WHERE expires_at < CURRENT_TIMESTAMP
    RETURNING warehouse_stock_id, quantity
), updated_stock AS (
    UPDATE warehouse_stock
    SET available_quantity = available_quantity + er.quantity,
        reserved_quantity = reserved_quantity - er.quantity,
        updated_at = CURRENT_TIMESTAMP
    FROM expired_reservations er
    WHERE warehouse_stock.id = er.warehouse_stock_id
    RETURNING warehouse_stock.created_at, warehouse_stock.updated_at, warehouse_stock.deleted_at, warehouse_stock.id, warehouse_stock.warehouse_id, warehouse_stock.product_id, warehouse_stock.available_quantity, warehouse_stock.reserved_quantity
)
INSERT INTO stock_history (warehouse_id, product_id, quantity_change, operation_type)
SELECT warehouse_id, product_id, quantity, 'release'
FROM updated_stock
RETURNING (SELECT created_at, updated_at, deleted_at, id, warehouse_id, product_id, available_quantity, reserved_quantity FROM updated_stock)
`

func (q *Queries) CancelExpiredReservations(ctx context.Context) ([]pgtype.Timestamp, error) {
	rows, err := q.db.Query(ctx, cancelExpiredReservations)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []pgtype.Timestamp{}
	for rows.Next() {
		var created_at pgtype.Timestamp
		if err := rows.Scan(&created_at); err != nil {
			return nil, err
		}
		items = append(items, created_at)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const confirmReservation = `-- name: ConfirmReservation :one
WITH removed_reservation AS (
    DELETE FROM stock_reservations
    WHERE stock_reservations.id = $1 AND stock_reservations.order_id = $2
    RETURNING warehouse_stock_id, quantity
), updated_stock AS (
    UPDATE warehouse_stock
    SET reserved_quantity = reserved_quantity - (SELECT quantity FROM removed_reservation),
        updated_at = CURRENT_TIMESTAMP
    WHERE warehouse_stock.id = (SELECT warehouse_stock_id FROM removed_reservation)
    RETURNING warehouse_stock.created_at, warehouse_stock.updated_at, warehouse_stock.deleted_at, warehouse_stock.id, warehouse_stock.warehouse_id, warehouse_stock.product_id, warehouse_stock.available_quantity, warehouse_stock.reserved_quantity
)
INSERT INTO stock_history (warehouse_id, product_id, quantity_change, operation_type)
SELECT warehouse_id, product_id, -(SELECT quantity FROM removed_reservation), 'confirmed'
FROM updated_stock
RETURNING (SELECT created_at, updated_at, deleted_at, id, warehouse_id, product_id, available_quantity, reserved_quantity FROM updated_stock)
`

type ConfirmReservationParams struct {
	ID      uuid.UUID `json:"id"`
	OrderID uuid.UUID `json:"orderId"`
}

func (q *Queries) ConfirmReservation(ctx context.Context, arg ConfirmReservationParams) (pgtype.Timestamp, error) {
	row := q.db.QueryRow(ctx, confirmReservation, arg.ID, arg.OrderID)
	var created_at pgtype.Timestamp
	err := row.Scan(&created_at)
	return created_at, err
}

const createWarehouse = `-- name: CreateWarehouse :one
INSERT INTO warehouses (
    name, location_id, location_name, capacity
) VALUES (
    $1, $2, $3, $4
) RETURNING created_at, updated_at, deleted_at, id, name, location_id, location_name, capacity
`

type CreateWarehouseParams struct {
	Name         string    `json:"name"`
	LocationID   uuid.UUID `json:"locationId"`
	LocationName string    `json:"locationName"`
	Capacity     int32     `json:"capacity"`
}

func (q *Queries) CreateWarehouse(ctx context.Context, arg CreateWarehouseParams) (Warehouse, error) {
	row := q.db.QueryRow(ctx, createWarehouse,
		arg.Name,
		arg.LocationID,
		arg.LocationName,
		arg.Capacity,
	)
	var i Warehouse
	err := row.Scan(
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.ID,
		&i.Name,
		&i.LocationID,
		&i.LocationName,
		&i.Capacity,
	)
	return i, err
}

const insertNewStock = `-- name: InsertNewStock :one
INSERT INTO warehouse_stock (
    warehouse_id, product_id, available_quantity, reserved_quantity
) VALUES (
    $1, $2, $3, $4
) RETURNING created_at, updated_at, deleted_at, id, warehouse_id, product_id, available_quantity, reserved_quantity
`

type InsertNewStockParams struct {
	WarehouseID       uuid.UUID `json:"warehouseId"`
	ProductID         uuid.UUID `json:"productId"`
	AvailableQuantity int32     `json:"availableQuantity"`
	ReservedQuantity  int32     `json:"reservedQuantity"`
}

func (q *Queries) InsertNewStock(ctx context.Context, arg InsertNewStockParams) (WarehouseStock, error) {
	row := q.db.QueryRow(ctx, insertNewStock,
		arg.WarehouseID,
		arg.ProductID,
		arg.AvailableQuantity,
		arg.ReservedQuantity,
	)
	var i WarehouseStock
	err := row.Scan(
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.ID,
		&i.WarehouseID,
		&i.ProductID,
		&i.AvailableQuantity,
		&i.ReservedQuantity,
	)
	return i, err
}

const reserveStock = `-- name: ReserveStock :one
WITH updated_stock AS (
    UPDATE warehouse_stock
    SET available_quantity = available_quantity - $1,
        reserved_quantity = reserved_quantity + $1,
        updated_at = CURRENT_TIMESTAMP
    WHERE warehouse_stock.warehouse_id = $2 AND warehouse_stock.product_id = $3 AND warehouse_stock.available_quantity >= $1
    RETURNING warehouse_stock.created_at, warehouse_stock.updated_at, warehouse_stock.deleted_at, warehouse_stock.id, warehouse_stock.warehouse_id, warehouse_stock.product_id, warehouse_stock.available_quantity, warehouse_stock.reserved_quantity
), inserted_reservation AS (
    INSERT INTO stock_reservations (warehouse_stock_id, order_id, quantity, expires_at)
    SELECT id, $4, $1, CURRENT_TIMESTAMP + INTERVAL $5
    FROM updated_stock
    RETURNING created_at, updated_at, deleted_at, id, warehouse_stock_id, order_id, quantity
)
INSERT INTO stock_history (warehouse_id, product_id, quantity_change, operation_type)
SELECT warehouse_id, product_id, -$1, 'reserve'
FROM updated_stock
RETURNING (SELECT created_at, updated_at, deleted_at, id, warehouse_stock_id, order_id, quantity FROM inserted_reservation)
`

type ReserveStockParams struct {
	Column1     interface{}     `json:"column1"`
	WarehouseID uuid.UUID       `json:"warehouseId"`
	ProductID   uuid.UUID       `json:"productId"`
	OrderID     uuid.UUID       `json:"orderId"`
	Column5     pgtype.Interval `json:"column5"`
}

func (q *Queries) ReserveStock(ctx context.Context, arg ReserveStockParams) (pgtype.Timestamp, error) {
	row := q.db.QueryRow(ctx, reserveStock,
		arg.Column1,
		arg.WarehouseID,
		arg.ProductID,
		arg.OrderID,
		arg.Column5,
	)
	var created_at pgtype.Timestamp
	err := row.Scan(&created_at)
	return created_at, err
}

const updateStock = `-- name: UpdateStock :one
UPDATE warehouse_stock
SET available_quantity = $1, reserved_quantity = $2
WHERE warehouse_id = $3 AND product_id = $4
RETURNING created_at, updated_at, deleted_at, id, warehouse_id, product_id, available_quantity, reserved_quantity
`

type UpdateStockParams struct {
	AvailableQuantity int32     `json:"availableQuantity"`
	ReservedQuantity  int32     `json:"reservedQuantity"`
	WarehouseID       uuid.UUID `json:"warehouseId"`
	ProductID         uuid.UUID `json:"productId"`
}

func (q *Queries) UpdateStock(ctx context.Context, arg UpdateStockParams) (WarehouseStock, error) {
	row := q.db.QueryRow(ctx, updateStock,
		arg.AvailableQuantity,
		arg.ReservedQuantity,
		arg.WarehouseID,
		arg.ProductID,
	)
	var i WarehouseStock
	err := row.Scan(
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.ID,
		&i.WarehouseID,
		&i.ProductID,
		&i.AvailableQuantity,
		&i.ReservedQuantity,
	)
	return i, err
}

const updateWarehouse = `-- name: UpdateWarehouse :one
UPDATE warehouses
SET name = $1, location_id = $2, location_name = $3, capacity = $4
WHERE id = $5
RETURNING created_at, updated_at, deleted_at, id, name, location_id, location_name, capacity
`

type UpdateWarehouseParams struct {
	Name         string    `json:"name"`
	LocationID   uuid.UUID `json:"locationId"`
	LocationName string    `json:"locationName"`
	Capacity     int32     `json:"capacity"`
	ID           uuid.UUID `json:"id"`
}

func (q *Queries) UpdateWarehouse(ctx context.Context, arg UpdateWarehouseParams) (Warehouse, error) {
	row := q.db.QueryRow(ctx, updateWarehouse,
		arg.Name,
		arg.LocationID,
		arg.LocationName,
		arg.Capacity,
		arg.ID,
	)
	var i Warehouse
	err := row.Scan(
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.ID,
		&i.Name,
		&i.LocationID,
		&i.LocationName,
		&i.Capacity,
	)
	return i, err
}