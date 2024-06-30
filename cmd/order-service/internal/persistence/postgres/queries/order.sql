-- name: GetOrders :many
SELECT 
    o.id AS order_id,
    o.customer_id,
    o.order_date,
    o.status,
    o.total,
    oi.id AS order_item_id,
    oi.product_id,
    oi.quantity,
    oi.price
FROM orders o
JOIN order_items oi ON o.id = oi.order_id
WHERE o.id = $1;

-- name: InsertOrder :one
INSERT INTO orders (id, customer_id, order_date, status, total)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, customer_id, order_date, status, total;

-- name: InsertOrderItem :exec
INSERT INTO order_items (order_id, product_id, quantity, price)
SELECT 
    unnest($1::uuid[]), 
    unnest($2::int[]), 
    unnest($3::int[]), 
    unnest($4::numeric[]);

-- name: UpdateOrder :one
UPDATE orders
SET status = $2, total = $3
WHERE id = $1
RETURNING id, customer_id, order_date, status, total;

-- name: UpdateOrderItem :exec
UPDATE order_items
SET product_id = $3, quantity = $4, price = $5
WHERE id = $1 AND order_id = $2;
