DROP TYPE IF EXISTS TRANSACTION_TYPE;
CREATE TYPE TRANSACTION_TYPE as ENUM ('DEPOSIT', 'WITHDRAW', 'INIT');

CREATE TABLE transactions
(
    id           uuid        DEFAULT uuid_generate_v4() PRIMARY KEY,
    order_number VARCHAR(255),
    user_id      uuid REFERENCES users (id)                    NOT NULL,
    old_balance  DECIMAL(32, 18)                               NOT NULL,
    change       DECIMAL(32, 18)                               NOT NULL,
    new_balance  DECIMAL(32, 18)                               NOT NULL,
    type         TRANSACTION_TYPE                              NOT NULL,
    created_at   TIMESTAMP DEFAULT NOW(),
    CONSTRAINT transactions_order_number_uniq UNIQUE NULLS DISTINCT(order_number)
)