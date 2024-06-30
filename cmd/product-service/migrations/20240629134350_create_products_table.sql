-- migrate:up

-- Create the products table
CREATE TABLE products (
    LIKE template_table INCLUDING ALL,
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL
);

-- Create an index on the name column
CREATE INDEX idx_products_name ON products(name);

-- Create the product_stocks table
CREATE TABLE product_stocks (
    LIKE template_table INCLUDING ALL,
    product_id INTEGER NOT NULL,
    warehouse_id INTEGER NOT NULL,
    stock_level INTEGER DEFAULT 0,
    CONSTRAINT fk_product
        FOREIGN KEY(product_id) 
        REFERENCES products(id)
        ON DELETE CASCADE,
    PRIMARY KEY (product_id, warehouse_id)
);

-- migrate:down

-- Drop the product_stocks table
DROP TABLE IF EXISTS product_stocks;

-- Drop the products table
DROP TABLE IF EXISTS products;

-- Drop the index on the name column
DROP INDEX IF EXISTS idx_products_name;
