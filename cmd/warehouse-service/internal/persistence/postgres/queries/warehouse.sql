-- name: CreateWarehouse :one
INSERT INTO warehouses (
    name, location_id, location_name, capacity
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: UpdateWarehouse :one
UPDATE warehouses
SET name = $1, location_id = $2, location_name = $3, capacity = $4
WHERE id = $5
RETURNING *;

-- name: InsertNewStock :one
INSERT INTO warehouse_stock (
    warehouse_id, product_id, available_quantity, reserved_quantity
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: UpdateStock :one
UPDATE warehouse_stock
SET available_quantity = $1, reserved_quantity = $2
WHERE warehouse_id = $3 AND product_id = $4
RETURNING *;

-- name: ReserveStock :one
WITH updated_stock AS (
    UPDATE warehouse_stock
    SET available_quantity = available_quantity - $1,
        reserved_quantity = reserved_quantity + $1,
        updated_at = CURRENT_TIMESTAMP
    WHERE warehouse_stock.warehouse_id = $2 AND warehouse_stock.product_id = $3 AND warehouse_stock.available_quantity >= $1
    RETURNING warehouse_stock.*
), inserted_reservation AS (
    INSERT INTO stock_reservations (warehouse_stock_id, order_id, quantity, expires_at)
    SELECT id, $4, $1, CURRENT_TIMESTAMP + INTERVAL $5
    FROM updated_stock
    RETURNING *
)
INSERT INTO stock_history (warehouse_id, product_id, quantity_change, operation_type)
SELECT warehouse_id, product_id, -$1, 'reserve'
FROM updated_stock
RETURNING (SELECT * FROM inserted_reservation);

-- name: ConfirmReservation :one
WITH removed_reservation AS (
    DELETE FROM stock_reservations
    WHERE stock_reservations.id = $1 AND stock_reservations.order_id = $2
    RETURNING warehouse_stock_id, quantity
), updated_stock AS (
    UPDATE warehouse_stock
    SET reserved_quantity = reserved_quantity - (SELECT quantity FROM removed_reservation),
        updated_at = CURRENT_TIMESTAMP
    WHERE warehouse_stock.id = (SELECT warehouse_stock_id FROM removed_reservation)
    RETURNING warehouse_stock.*
)
INSERT INTO stock_history (warehouse_id, product_id, quantity_change, operation_type)
SELECT warehouse_id, product_id, -(SELECT quantity FROM removed_reservation), 'confirmed'
FROM updated_stock
RETURNING (SELECT * FROM updated_stock);

-- name: CancelExpiredReservations :many
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
    RETURNING warehouse_stock.*
)
INSERT INTO stock_history (warehouse_id, product_id, quantity_change, operation_type)
SELECT warehouse_id, product_id, quantity, 'release'
FROM updated_stock
RETURNING (SELECT * FROM updated_stock);
