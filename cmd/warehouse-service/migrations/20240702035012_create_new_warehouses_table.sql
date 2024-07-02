-- migrate:up
CREATE TABLE warehouses (
    LIKE template_table INCLUDING ALL,
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    location_id UUID NOT NULL,
    location_name VARCHAR(255) NOT NULL,
    capacity INTEGER NOT NULL
);

-- migrate:down
DROP TABLE warehouses;
