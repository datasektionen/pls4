begin;

insert into users values
    ('vakant'),
    ('turetek');

insert into roles values
    ('ordf', 'Ordförande', 'Förför ord'),
    ('dsys', 'Systemansvarig', 'Ansvarar över System'),
    ('dfunkt', 'dFunkt', 'Alla dFunktionärer'),
    ('ior', 'Informationsorganet', 'Informerar om organ');

insert into roles_users (role_id, kth_id, comment, modified_by, start_date, end_date)
values (
    'ordf',
    'turetek',
    'Vald på Val-SM 2023',
    'turetek',
    '2023-04-01',
    '2024-03-31'
), (
    'dsys',
    'turetek',
    'Vald på Val-SM 2023',
    'turetek',
    '2023-07-01',
    '2024-12-31'
), (
    'ior',
    'vakant',
    '',
    'turetek',
    '2023-09-09',
    '2024-09-09'
);

insert into roles_roles (superrole_id, subrole_id) values
    ('dfunkt', 'ordf'),
    ('dfunkt', 'dsys'),
    ('ior', 'dsys');

insert into permissions (id, system, name)
values ('0038b4e6-0027-45ac-ae26-9534e8ffc9bd', 'calypso', 'create');

insert into permissions (system, name)
values ('jml', 'admin');

insert into roles_permissions
values ('dfunkt', '0038b4e6-0027-45ac-ae26-9534e8ffc9bd');

insert into api_tokens (id, description, expires_at)
values (
    '689ad7d4-74d4-4e42-9836-451f1045f117',
    'test',
    '2024-01-01'
);

insert into api_tokens_permissions (api_token_id, permission_id)
values (
    '689ad7d4-74d4-4e42-9836-451f1045f117',
    '0038b4e6-0027-45ac-ae26-9534e8ffc9bd'
);

commit;
