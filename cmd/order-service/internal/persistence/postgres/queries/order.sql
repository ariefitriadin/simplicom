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

-- name: GetOrder :one
SELECT * FROM orders WHERE id = $1;

-- name: InsertOrder :one
INSERT INTO orders (id, customer_id, order_date, status, total)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, customer_id, order_date, status, total;

-- name: InsertOrderItem :exec
INSERT INTO order_items (order_id, product_id, product_name, quantity, price)
SELECT 
    unnest($1::uuid[]), 
    unnest($2::int[]), 
    unnest($3::text[]),
    unnest($4::int[]), 
    unnest($5::numeric[]);

-- name: UpdateOrder :one
UPDATE orders
SET status = $2, total = $3
WHERE id = $1
RETURNING id, customer_id, order_date, status, total;

-- name: UpdateOrderItems :exec
WITH updates AS (
    SELECT 
        unnest($1::uuid[]) AS id,
        unnest($2::uuid[]) AS order_id,
        unnest($3::int[]) AS product_id,
        unnest($4::text[]) AS product_name,
        unnest($5::int[]) AS quantity,
        unnest($6::numeric[]) AS price
)
UPDATE order_items oi
SET 
    product_id = u.product_id,
    product_name = u.product_name,
    quantity = u.quantity,
    price = u.price
FROM updates u
WHERE oi.id = u.id AND oi.order_id = u.order_id;
