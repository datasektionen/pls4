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

create table api_tokens (
    id          uuid primary key default gen_random_uuid(),
    secret      uuid unique not null default gen_random_uuid(),
    description text not null,

    created_at   timestamp not null default now(),
    expires_at   timestamp,
    last_used_at timestamp
);

create table systems (
    id text not null primary key check(id ~ '^[a-z0-9]+(-[a-z0-9]+)*$')
);

create table permissions (
    system_id  text not null,
    id         text not null check(id ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    has_scope  bool not null,

    foreign key (system_id) references systems (id),
    primary key (system_id, id)
);

create table permission_instances (
    id            uuid primary key default gen_random_uuid(),
    system_id     text not null,
    permission_id text not null,
    scope         text check(scope != ''),

    foreign key (system_id)                references systems (id),
    foreign key (system_id, permission_id) references permissions (system_id, id)
);

create table roles_permissions (
    permission_instance_id uuid primary key,
    role_id                text not null,

    foreign key (permission_instance_id) references permission_instances (id) on delete cascade,
    foreign key (role_id)                references roles                (id)
);

create table api_tokens_permissions (
    permission_instance_id uuid primary key,
    api_token_id           uuid not null,

    foreign key (permission_instance_id) references permission_instances (id),
    foreign key (api_token_id)           references api_tokens           (id)
);

create table sessions (
    id           uuid primary key default gen_random_uuid(),
    kth_id       text      not null,
    display_name text      not null,
    last_used_at timestamp not null
);

insert into systems (id) values ('pls');
insert into permissions (system_id, id, has_scope) values
    ('pls', 'create-role', false),
    ('pls', 'system', true),
    ('pls', 'role', true),
    ('pls', 'manage-systems', false);
