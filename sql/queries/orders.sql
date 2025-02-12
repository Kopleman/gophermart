-- name: GetOrderByNumber :one
SELECT id, user_id, order_number, status, accrual, created_at, updated_at, deleted_at FROM orders WHERE order_number=$1 LIMIT 1;

-- name: CreateOrder :one
INSERT INTO orders (user_id, order_number, status)
VALUES ($1, $2,  'NEW')
RETURNING id, user_id, order_number, status, accrual, created_at, updated_at, deleted_at;

-- name: PutOrderForProcessing :one
INSERT INTO orders_to_process (order_number, process_status)
VALUES ($1, 'NEW')
RETURNING order_number, process_status, created_at, updated_at, deleted_at;

-- name: PickOrdersToProcess :many
UPDATE orders_to_process SET process_status = 'START_PROCESSING', updated_at = now()
WHERE order_number IN (SELECT order_number
    FROM orders_to_process
    WHERE process_status = 'NEW' ORDER BY created_at ASC
    LIMIT $1)
RETURNING order_number, process_status, created_at, updated_at, deleted_at;

-- name: GetStartProcessingOrders :many
SELECT order_number, process_status, created_at, updated_at, deleted_at from orders_to_process WHERE process_status = 'START_PROCESSING' ORDER BY created_at ASC;

-- name: GetUserOrders :many
SELECT id, user_id, order_number, status, accrual, created_at, updated_at, deleted_at FROM orders WHERE user_id=$1 ORDER BY created_at DESC;

-- name: UpdateOrder :one
UPDATE orders SET status = $1, accrual = $2, updated_at = now()
WHERE order_number = $3
RETURNING id, user_id, order_number, status, accrual, created_at, updated_at, deleted_at;

-- name: UpdateOrderToProcess :one
UPDATE orders_to_process SET process_status = $1, updated_at = now()
WHERE order_number = $2
RETURNING order_number, process_status, created_at, updated_at, deleted_at;

-- name: GetRegisteredProcessingOrders :many
SELECT order_number, process_status, created_at, updated_at, deleted_at from orders_to_process WHERE process_status = 'REGISTERED' ORDER BY created_at ASC LIMIT $1;