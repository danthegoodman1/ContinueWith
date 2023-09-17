
-- +migrate Up

create table clients (
    id text not null,
    suspended bool not null default false,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (id)
)
;

-- scopes that the provider has defined
create table scopes (
    id text not null,
    description text,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key(id)
)
;

create table authorization_codes (
    id text not null,
    user_id text not null,
    scopes text[] not null default '{}',
    expires timestamptz not null,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key(id)
)
;

-- +migrate Down
drop table clients;
drop table scopes;
drop table authorization_codes;