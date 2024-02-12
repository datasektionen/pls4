create table roles (
    id           text primary key check(id ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    display_name text not null,
    description  text not null
);

create table roles_users (
    id       uuid primary key default gen_random_uuid(),
    role_id  text not null,
    kth_id   text not null,

    modified_by text      not null,
    modified_at timestamp not null default now(),
    start_date  date      not null default now(),
    end_date    date      not null,

    foreign key (role_id) references roles (id)
);

create table roles_roles (
    superrole_id text not null,
    subrole_id   text not null,

    foreign key (superrole_id) references roles (id),
    foreign key (subrole_id)   references roles (id),
    primary key (superrole_id, subrole_id)
);

create table roles_permissions (
    role_id       text not null,
    system        text not null,
    permission    text not null check(permission ~ '^\*|[a-z0-9]+(-[a-z0-9]+)*(-\*)?$'),

    foreign key (role_id) references roles (id),
    primary key (role_id, system, permission)
);

create table api_tokens (
    id          uuid primary key default gen_random_uuid(),
    secret      uuid unique not null default gen_random_uuid(),
    description text not null,

    created_at   timestamp not null default now(),
    expires_at   timestamp not null,
    last_used_at timestamp
);

create table api_tokens_permissions (
    api_token_id  uuid not null,
    system        text not null,
    permission    text not null check(permission ~ '^\*|[a-z0-9]+(-[a-z0-9]+)*(-\*)?$'),

    foreign key (api_token_id) references api_tokens (id),
    primary key (api_token_id, system, permission)
);

create table sessions (
    id           uuid primary key default gen_random_uuid(),
    kth_id       text      not null,
    display_name text      not null,
    last_used_at timestamp not null
)
