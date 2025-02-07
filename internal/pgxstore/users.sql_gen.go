// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: users.sql

package pgxstore

import (
	"context"

	"github.com/google/uuid"
)

const CreateUser = `-- name: CreateUser :one
INSERT INTO users (login, password_hash, created_at)
VALUES ($1, $2,  now())
RETURNING id, login, password_hash, created_at, updated_at, deleted_at
`

type CreateUserParams struct {
	Login        string `db:"login" json:"login"`
	PasswordHash string `db:"password_hash" json:"password_hash"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (*User, error) {
	row := q.db.QueryRow(ctx, CreateUser, arg.Login, arg.PasswordHash)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Login,
		&i.PasswordHash,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return &i, err
}

const GetAllUsers = `-- name: GetAllUsers :many
SELECT id, login, password_hash, created_at, updated_at, deleted_at FROM users ORDER BY created_at ASC
`

func (q *Queries) GetAllUsers(ctx context.Context) ([]*User, error) {
	rows, err := q.db.Query(ctx, GetAllUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Login,
			&i.PasswordHash,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const GetUserById = `-- name: GetUserById :one
SELECT id, login, password_hash, created_at, updated_at, deleted_at FROM users WHERE id=$1 LIMIT 1
`

func (q *Queries) GetUserById(ctx context.Context, id uuid.UUID) (*User, error) {
	row := q.db.QueryRow(ctx, GetUserById, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Login,
		&i.PasswordHash,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return &i, err
}

const GetUserByLogin = `-- name: GetUserByLogin :one
SELECT id, login, password_hash, created_at, updated_at, deleted_at FROM users WHERE login=$1 LIMIT 1
`

func (q *Queries) GetUserByLogin(ctx context.Context, login string) (*User, error) {
	row := q.db.QueryRow(ctx, GetUserByLogin, login)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Login,
		&i.PasswordHash,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return &i, err
}
