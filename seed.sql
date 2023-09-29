begin;

insert into roles values
    ('ordf', 'Ordförande', 'Förför ord'),
    ('dsys', 'Systemansvarig', 'Ansvarar över System'),
    ('dfunkt', 'dFunkt', 'Alla dFunktionärer'),
    ('kassor', 'Kassör', 'Cash yo'),
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
    '2023-06-27',
    '2024-06-27'
), (
    'kassor',
    'snoren',
    'Vald på Bar-SM 2019',
    'turetek',
    '2019-06-27',
    '2026-06-27'
);

insert into roles_roles (superrole_id, subrole_id) values
    ('dfunkt', 'ordf'),
    ('dfunkt', 'dsys'),
    ('ior', 'dsys');

insert into roles_permissions (role_id, system, permission) values
    ('dfunkt', 'calypso', 'create'),
    ('kassor', 'cashflow', 'attest-*'),
    ('kassor', 'cashflow', 'confirm'),
    ('dsys', 'pls', 'create-role'),
    ('dsys', 'pls', '*');

insert into api_tokens (id, description, expires_at)
values ('689ad7d4-74d4-4e42-9836-451f1045f117', 'test', '2024-01-01');

insert into api_tokens_permissions (api_token_id, system, permission)
values ('689ad7d4-74d4-4e42-9836-451f1045f117', 'calypso', 'create');

commit;
