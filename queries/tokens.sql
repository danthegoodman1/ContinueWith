-- name: InsertRefreshToken :exec
insert into refresh_tokens (
    id
    , client_id
    , user_id
    , scopes
    , expires
) values (
    @id
    , @client_id
    , @user_id
    , @scopes
    , @expires
)
;

-- name: InsertAccessToken :exec
insert into access_tokens (
    id
    , client_id
    , refresh_token
    , user_id
    , scopes
    , expires
) values (
    @id
    , @client_id
    , @refresh_token
    , @user_id
    , @scopes
    , @expires
)
;

-- name: SelectValidAccessToken :one
select *
from access_tokens
where id = $1
and expires > now()
and revoked = false
;

-- name: SelectValidRefreshToken :one
-- Might be expired, if so need to regenerate and mark as revoked
select *
from refresh_tokens
where id = $1
and expires > now()
;

-- name: RevokeRefreshToken :exec
update refresh_tokens
set revoked = true
where id = $1
;

-- name: RevokeAccessToken :exec
update access_tokens
set revoked = true
where id = $1
;

-- name: ListRefreshTokensByUserID :many
select *
from refresh_tokens
where user_id = $1
;

-- name: ListAccessTokensByUserID :many
select *
from access_tokens
where user_id = $1
;