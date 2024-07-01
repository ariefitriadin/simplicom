-- migrate:up
-- Create the orders table
CREATE TABLE orders (
    LIKE template_table INCLUDING ALL,
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    customer_id UUID NOT NULL,
    order_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL,
    total DECIMAL(10, 2) NOT NULL
);

-- migrate:down
-- Drop the orders table
DROP TABLE IF EXISTS orders;
