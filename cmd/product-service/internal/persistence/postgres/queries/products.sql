-- name: GetAllProductsWithStock :many
SELECT 
    p.id,
    p.name,
    p.description,
    p.price,
    p.createdAt,
    p.updatedAt,
    ps.warehouse_id,
    ps.stock_level
FROM 
    products p
LEFT JOIN 
    product_stocks ps ON p.id = ps.product_id
LIMIT $1 OFFSET $2;


-- name: InsertProduct :one
INSERT INTO products (name, description, price)
VALUES ($1, $2, $3)
RETURNING id, name, description, price, createdAt, updatedAt;

-- name: InsertProductStock :one
INSERT INTO product_stocks (product_id, warehouse_id, stock_level)
VALUES ($1, $2, $3)
RETURNING product_id, warehouse_id, stock_level;

-- name: UpdateProductStock :exec
UPDATE product_stocks
SET stock_level = $1, warehouse_id = $2, updatedAt = NOW()
WHERE product_id = $3 AND warehouse_id = $4;

-- name: GetProductByID :one
SELECT 
    p.id,
    p.name,
    p.description,
    p.price,
    ps.warehouse_id,
    ps.stock_level
FROM 
    products p
LEFT JOIN 
    product_stocks ps ON p.id = ps.product_id
WHERE 
    p.id = $1;

-- name: DeleteProductStock :exec
DELETE FROM product_stocks
WHERE product_id = $1 AND warehouse_id = $2;
