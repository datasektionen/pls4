create table users (
    kth_id text primary key
);

create table roles (
    id           text primary key,
    display_name text not null,
    description  text not null
);

create table roles_users (
    id       uuid primary key default gen_random_uuid(),
    role_id  text not null,
    kth_id   text not null,

    comment     text      not null,
    modified_by text      not null,
    modified_at timestamp not null default now(),
    start_date  timestamp not null default now(),
    end_date    timestamp not null,

    foreign key (role_id)     references roles (id),
    foreign key (kth_id)      references users (kth_id),
    foreign key (modified_by) references users (kth_id)
);

create table roles_roles (
    superrole_id text not null,
    subrole_id   text not null,

    foreign key (superrole_id) references roles (id),
    foreign key (subrole_id)   references roles (id),
    primary key (superrole_id, subrole_id)
);

create table permissions (
    id     uuid primary key default gen_random_uuid(),
    system text not null,
    name   text not null
);

create table roles_permissions (
    role_id       text not null,
    permission_id uuid not null,

    foreign key (permission_id) references permissions (id)
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
    permission_id uuid not null,

    foreign key (api_token_id)  references api_tokens (id),
    foreign key (permission_id) references permissions (id),
    primary key (api_token_id, permission_id)
);
