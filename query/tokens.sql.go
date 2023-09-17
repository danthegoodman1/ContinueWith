// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1
// source: tokens.sql

package query

import (
	"context"
	"time"
)

const insertAccessToken = `-- name: InsertAccessToken :exec
insert into access_tokens (
    id
    , client_id
    , refresh_token
    , user_id
    , scopes
    , expires
) values (
    $1
    , $2
    , $3
    , $4
    , $5
    , $6
)
`

type InsertAccessTokenParams struct {
	ID           string
	ClientID     string
	RefreshToken string
	UserID       string
	Scopes       []string
	Expires      time.Time
}

func (q *Queries) InsertAccessToken(ctx context.Context, arg InsertAccessTokenParams) error {
	_, err := q.db.Exec(ctx, insertAccessToken,
		arg.ID,
		arg.ClientID,
		arg.RefreshToken,
		arg.UserID,
		arg.Scopes,
		arg.Expires,
	)
	return err
}

const insertRefreshToken = `-- name: InsertRefreshToken :exec
insert into refresh_tokens (
    id
    , client_id
    , user_id
    , scopes
    , expires
) values (
    $1
    , $2
    , $3
    , $4
    , $5
)
`

type InsertRefreshTokenParams struct {
	ID       string
	ClientID string
	UserID   string
	Scopes   []string
	Expires  time.Time
}

func (q *Queries) InsertRefreshToken(ctx context.Context, arg InsertRefreshTokenParams) error {
	_, err := q.db.Exec(ctx, insertRefreshToken,
		arg.ID,
		arg.ClientID,
		arg.UserID,
		arg.Scopes,
		arg.Expires,
	)
	return err
}

const listAccessTokensByUserID = `-- name: ListAccessTokensByUserID :many
select id, client_id, refresh_token, user_id, scopes, expires, revoked, created, updated
from access_tokens
where user_id = $1
`

func (q *Queries) ListAccessTokensByUserID(ctx context.Context, userID string) ([]AccessToken, error) {
	rows, err := q.db.Query(ctx, listAccessTokensByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AccessToken
	for rows.Next() {
		var i AccessToken
		if err := rows.Scan(
			&i.ID,
			&i.ClientID,
			&i.RefreshToken,
			&i.UserID,
			&i.Scopes,
			&i.Expires,
			&i.Revoked,
			&i.Created,
			&i.Updated,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listRefreshTokensByUserID = `-- name: ListRefreshTokensByUserID :many
select id, client_id, user_id, scopes, expires, revoked, created, updated
from refresh_tokens
where user_id = $1
`

func (q *Queries) ListRefreshTokensByUserID(ctx context.Context, userID string) ([]RefreshToken, error) {
	rows, err := q.db.Query(ctx, listRefreshTokensByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RefreshToken
	for rows.Next() {
		var i RefreshToken
		if err := rows.Scan(
			&i.ID,
			&i.ClientID,
			&i.UserID,
			&i.Scopes,
			&i.Expires,
			&i.Revoked,
			&i.Created,
			&i.Updated,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const revokeAccessToken = `-- name: RevokeAccessToken :exec
update access_tokens
set revoked = true
where id = $1
`

func (q *Queries) RevokeAccessToken(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, revokeAccessToken, id)
	return err
}

const revokeRefreshToken = `-- name: RevokeRefreshToken :exec
update refresh_tokens
set revoked = true
where id = $1
`

func (q *Queries) RevokeRefreshToken(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, revokeRefreshToken, id)
	return err
}

const selectValidAccessToken = `-- name: SelectValidAccessToken :one
select id, client_id, refresh_token, user_id, scopes, expires, revoked, created, updated
from access_tokens
where id = $1
and expires > now()
and revoked = false
`

func (q *Queries) SelectValidAccessToken(ctx context.Context, id string) (AccessToken, error) {
	row := q.db.QueryRow(ctx, selectValidAccessToken, id)
	var i AccessToken
	err := row.Scan(
		&i.ID,
		&i.ClientID,
		&i.RefreshToken,
		&i.UserID,
		&i.Scopes,
		&i.Expires,
		&i.Revoked,
		&i.Created,
		&i.Updated,
	)
	return i, err
}

const selectValidRefreshToken = `-- name: SelectValidRefreshToken :one
select id, client_id, user_id, scopes, expires, revoked, created, updated
from refresh_tokens
where id = $1
and expires > now()
and revoked = false
`

func (q *Queries) SelectValidRefreshToken(ctx context.Context, id string) (RefreshToken, error) {
	row := q.db.QueryRow(ctx, selectValidRefreshToken, id)
	var i RefreshToken
	err := row.Scan(
		&i.ID,
		&i.ClientID,
		&i.UserID,
		&i.Scopes,
		&i.Expires,
		&i.Revoked,
		&i.Created,
		&i.Updated,
	)
	return i, err
}