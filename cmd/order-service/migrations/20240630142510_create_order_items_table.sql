-- migrate:up

-- Create the order_items table
CREATE TABLE order_items (
    LIKE template_table INCLUDING ALL,
    id SERIAL PRIMARY KEY,
    order_id UUID NOT NULL,
    product_id INTEGER NOT NULL,
    product_name VARCHAR(250) NOT NULL,
    quantity INTEGER NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    CONSTRAINT fk_order
        FOREIGN KEY(order_id) 
        REFERENCES orders(id)
        ON DELETE CASCADE
);

-- migrate:down

-- Drop the order_items table
DROP TABLE IF EXISTS order_items;