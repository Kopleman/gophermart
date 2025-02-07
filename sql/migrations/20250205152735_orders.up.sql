DROP TYPE IF EXISTS ORDER_STATUS_TYPE;
CREATE TYPE ORDER_STATUS_TYPE AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS orders
(
    id           uuid                                DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id      uuid REFERENCES users (id) NOT NULL,
    order_number VARCHAR(255)               NOT NULL,
    status       ORDER_STATUS_TYPE          NOT NULL,
    accrual      DECIMAL(32, 18)            NOT NULL DEFAULT 0,
    created_at   TIMESTAMP                  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP                  NULL,
    deleted_at   TIMESTAMP                  NULL,
    CONSTRAINT order_number_uniq UNIQUE (order_number)
);

DROP TYPE IF EXISTS PROCESS_STATUS_TYPE;
CREATE TYPE PROCESS_STATUS_TYPE AS ENUM ('NEW', 'PROCESSING', 'PROCESSED');

CREATE TABLE IF NOT EXISTS orders_to_process
(
    order_number   VARCHAR(255) REFERENCES orders (order_number) NOT NULL,
    process_status PROCESS_STATUS_TYPE                           NOT NULL,
    created_at     TIMESTAMP                                     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP                                     NULL,
    deleted_at     TIMESTAMP                                     NULL,
    CONSTRAINT process_order_number_uniq UNIQUE (order_number)
);