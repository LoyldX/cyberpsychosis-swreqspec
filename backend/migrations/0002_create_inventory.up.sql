-- FR-INV-01, FR-INV-02, FR-INV-03, DOM-INV-02, CON-DB-01: schema สำหรับ inventory และประวัติ movement
CREATE TABLE inventory_items (
    id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name              TEXT        NOT NULL,
    item_type         TEXT        NOT NULL,
    quantity_on_hand  BIGINT      NOT NULL DEFAULT 0,
    reorder_point     BIGINT,
    deleted_at        TIMESTAMPTZ
);

CREATE TABLE inventory_movements (
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    inventory_item_id  BIGINT      NOT NULL REFERENCES inventory_items (id),
    movement_type      TEXT        NOT NULL,
    quantity           BIGINT      NOT NULL,
    responsible_person TEXT,
    purpose            TEXT,
    due_date           DATE,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
