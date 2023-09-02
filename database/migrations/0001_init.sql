create table users (
    kth_id text primary key
);

create table groups (
    id           text primary key,
    display_name text not null,
    description  text not null
);

create table groups_users (
    id       uuid primary key default gen_random_uuid(),
    group_id text not null,
    kth_id   text not null,

    comment     text      not null,
    modified_by text      not null,
    modified_at timestamp not null default now(),
    start_date  timestamp not null default now(),
    end_date    timestamp not null,

    foreign key (group_id)    references groups (id),
    foreign key (kth_id)      references users (kth_id),
    foreign key (modified_by) references users (kth_id)
);

create table groups_groups (
    supergroup_id text not null,
    subgroup_id   text not null,

    foreign key (supergroup_id) references groups (id),
    foreign key (subgroup_id)   references groups (id),
    primary key (supergroup_id, subgroup_id)
);

create table permissions (
    id     uuid primary key default gen_random_uuid(),
    system text not null,
    name   text not null
);

create table groups_permissions (
    group_id      text not null,
    permission_id uuid not null,

    foreign key (permission_id) references permissions (id)
);
