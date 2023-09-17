-- name: InsertAuthorizationCode :exec
insert into authorization_codes (
    id
    , user_id
    , client_id
    , scopes
    , expires
) values (
     @id
     , @user_id
     , @client_id
     , @scopes
     , @expires
 )
;

-- name: SelectAuthorizationCode :one
select *
from authorization_codes
where id = $1
;

-- name: DeleteAuthorizationCode :one
delete from authorization_codes
where id = $1
returning *
;