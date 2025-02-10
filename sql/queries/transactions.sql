-- name: CreateInitUserTransaction :one
INSERT INTO transactions (user_id, old_balance, change, new_balance, type, created_at)
VALUES ($1, 0,  0, 0, 'INIT', now())
RETURNING id, order_number, user_id, old_balance, change, new_balance, type, created_at;

-- name: GetLastUserTransaction :one
SELECT id, order_number, user_id, old_balance, change, new_balance, type, created_at FROM transactions WHERE user_id=$1 ORDER BY created_at DESC LIMIT 1;

-- name: GetTransactions :many
SELECT id, order_number, user_id, old_balance, change, new_balance, type, created_at
    FROM transactions as t
    WHERE
        (NOT @is_user_id::boolean OR t.user_id = @user_id) AND
        (NOT @is_type::boolean OR t.type = @type)
    ORDER BY created_at DESC;

-- name: GetUserWithdrawalsSum :one
SELECT COALESCE(SUM(change),0)::DECIMAL(32, 18) as withdrawals FROM transactions WHERE user_id=$1 and type='WITHDRAW';

-- name: CreateTransaction :one
INSERT INTO transactions (order_number, user_id, old_balance, change, new_balance, type, created_at)
VALUES ($1, $2,  $3, $4, $5, $6, now())
RETURNING id, order_number, user_id, old_balance, change, new_balance, type, created_at;