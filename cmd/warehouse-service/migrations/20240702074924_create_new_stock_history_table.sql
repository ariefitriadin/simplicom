-- migrate:up
CREATE TABLE stock_history (
    LIKE template_table including all,
    id SERIAL PRIMARY KEY,
    warehouse_id uuid NOT NULL,
    product_id uuid NOT NULL,
    quantity_change INTEGER NOT NULL,
    operation_type VARCHAR(50) NOT NULL,  -- add, remove, reserve, release, adjust, confirmed
    user_id uuid NOT NULL, --employee / pic 
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE CASCADE
);


-- migrate:down
DROP TABLE stock_history;