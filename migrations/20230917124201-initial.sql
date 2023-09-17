
-- +migrate Up

create table clients (
    id text not null,
    secret text not null,
    suspended bool not null default false,

    created timestamptz not null default now(),
    updated timestamptz not null default now(),
    primary key (id)
)
;

-- scopes that the provider has defined
create table scopes (
    id text not null,
    description text,

    created timestamptz not null default now(),
    updated timestamptz not null default now(),
    primary key(id)
)
;

create table authorization_codes (
    id text not null,
    client_id text not null references clients(id) on delete cascade ,
    user_id text not null,
    scopes text[] not null default '{}',
    expires timestamptz not null,

    created timestamptz not null default now(),
    updated timestamptz not null default now(),
    primary key(id)
)
;

create table refresh_tokens (
     id text not null,
     client_id text not null references clients(id) on delete cascade ,
     user_id text not null,
     scopes text[] not null default '{}', -- the max scopes this is verified for
     expires timestamptz not null,
     revoked bool not null default false,

     created timestamptz not null default now(),
     updated timestamptz not null default now(),
     primary key(id)
)
;

create index refresh_tokens_by_user_id on refresh_tokens(user_id);

create table access_tokens (
     id text not null,
     client_id text not null references clients(id) on delete cascade,
     refresh_token text not null,
     user_id text not null,
     scopes text[] not null default '{}', -- a subset of the refresh token scopes
     expires timestamptz not null,
     revoked bool not null default false,

     created timestamptz not null default now(),
     updated timestamptz not null default now(),
     primary key(id)
)
;
create index access_tokens_by_user_id on access_tokens(user_id);

-- +migrate Down
drop table clients;
drop table scopes;
drop table authorization_codes;
drop table refresh_tokens;
drop table access_tokens;