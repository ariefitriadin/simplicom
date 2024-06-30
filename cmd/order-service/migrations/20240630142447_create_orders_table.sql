-- migrate:up
-- Create the orders table
CREATE TABLE orders (
    LIKE template_table INCLUDING ALL,
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),  --generate ulid type
    customer_id INTEGER NOT NULL,
    order_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL,
    total DECIMAL(10, 2) NOT NULL
);

-- migrate:down
-- Drop the orders table
DROP TABLE IF EXISTS orders;
