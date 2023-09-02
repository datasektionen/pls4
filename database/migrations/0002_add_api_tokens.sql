create table api_tokens (
    id          uuid primary key default gen_random_uuid(),
    secret      uuid not null default gen_random_uuid(),
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
