begin;

insert into roles values
    ('ordf', 'Ordförande', 'Förför ord'),
    ('dsys', 'Systemansvarig', 'Ansvarar över System'),
    ('dfunkt', 'dFunkt', 'Alla dFunktionärer'),
    ('kassor', 'Kassör', 'Cash yo'),
    ('ior', 'Informationsorganet', 'Informerar om organ'),
    ('komm', 'Kommunikatör', 'Kommunicerar och ser till att folk kommunierar rätt'),
    ('drek', 'D-rektoratet', 'Styr över skit');

insert into roles_users (role_id, kth_id, modified_by, start_date, end_date)
values (
    'ordf',
    'adamsjo',
    'vakant',
    '2024-01-01',
    '2024-12-31'
), (
    'dsys',
    'mathm',
    'vakant',
    '2023-07-01',
    '2024-12-31'
), (
    'dsys',
    'turetek',
    'vakant',
    '2000-01-01',
    '2099-12-31'
), (
    'ior',
    'hermanka',
    'vakant',
    '2017-07-14',
    '2024-12-31'
), (
    'ior',
    'nilsmal',
    'vakant',
    '2023-07-14',
    '2024-12-31'
), (
    'kassor',
    'melvinj',
    'vakant',
    '2024-01-01',
    '2024-12-31'
), (
    'komm',
    'bwidman',
    'vakant',
    '2023-07-01',
    '2024-06-30'
);

insert into roles_roles (superrole_id, subrole_id) values
    ('drek', 'ordf'),
    ('drek', 'kassor'),
    ('dfunkt', 'drek'),
    ('dfunkt', 'dsys'),
    ('ior', 'komm'),
    ('ior', 'dsys');

insert into roles_permissions (role_id, system, permission) values
    ('dfunkt', 'calypso', 'post'),
    ('dfunkt', 'pico', 'custom'),
    ('dfunkt', 'hodis', 'membercheck'),
    ('kassor', 'cashflow', 'attest-*'),
    ('kassor', 'cashflow', 'confirm'),
    ('kassor', 'pls', 'system-cashflow'),
    ('drek', 'cashflow', 'see-all'),
    ('drek', 'pls', 'create-role'),
    ('dsys', 'pls', '*'),
    ('komm', 'pls', 'system-calypso');

insert into api_tokens (id, description, expires_at)
values ('689ad7d4-74d4-4e42-9836-451f1045f117', 'test', '2024-01-01');

insert into api_tokens_permissions (api_token_id, system, permission)
values ('689ad7d4-74d4-4e42-9836-451f1045f117', 'calypso', 'create');

commit;
