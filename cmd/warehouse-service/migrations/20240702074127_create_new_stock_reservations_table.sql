-- migrate:up
CREATE TABLE stock_reservations (
    LIKE template_table INCLUDING ALL,
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    warehouse_stock_id UUID NOT NULL,
    order_id UUID NOT NULL,
    quantity INTEGER NOT NULL,
    FOREIGN KEY (warehouse_stock_id) REFERENCES warehouse_stock(id) ON DELETE CASCADE
);

-- migrate:down
DROP TABLE stock_reservations;
