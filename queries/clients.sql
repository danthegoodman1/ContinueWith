-- name: SelectClient :one
select *
from clients
where id = $1
;