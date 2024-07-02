-- migrate:up
CREATE TABLE warehouse_stock (
    LIKE template_table INCLUDING ALL,
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    warehouse_id UUID NOT NULL,
    product_id UUID NOT NULL,
    available_quantity INTEGER NOT NULL,
    reserved_quantity INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE CASCADE,
    CONSTRAINT check_non_negative_quantities CHECK (available_quantity >= 0 AND reserved_quantity >= 0)
);

-- migrate:down
DROP TABLE warehouse_stock;
