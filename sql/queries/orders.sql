-- name: GetOrderByNumber :one
SELECT id, user_id, order_number, status, created_at, updated_at, deleted_at FROM orders WHERE order_number=$1 LIMIT 1;

-- name: CreateOrder :one
INSERT INTO orders (user_id, order_number, status)
VALUES ($1, $2,  'NEW')
RETURNING id, user_id, order_number, status, created_at, updated_at, deleted_at;

-- name: PutOrderForProcessing :one
INSERT INTO orders_to_process (order_id, process_status)
VALUES ($1, 'NEW')
RETURNING order_id, process_status, created_at, updated_at, deleted_at;