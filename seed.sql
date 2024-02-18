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

insert into systems (id) values
    ('calypso'),
    ('pico'),
    ('hodis'),
    ('cashflow');

insert into permissions (system_id, id, has_scope) values
    ('calypso', 'post', false),
    ('pico', 'custom', false),
    ('hodis', 'membercheck', false),
    ('cashflow', 'attest', true),
    ('cashflow', 'confirm', false),
    ('cashflow', 'see-all', false);

insert into permission_instances (id, system_id, permission_id, scope) values
    ('00000000-0000-0000-0000-000000000000', 'calypso', 'post', null),
    ('00000000-0000-0000-0000-000000000001', 'pico', 'custom', null),
    ('00000000-0000-0000-0000-000000000002', 'hodis', 'membercheck', null),
    ('00000000-0000-0000-0000-000000000003', 'cashflow', 'attest', '*'),
    ('00000000-0000-0000-0000-000000000004', 'cashflow', 'confirm', null),
    ('00000000-0000-0000-0000-000000000005', 'cashflow', 'see-all', null),
    ('00000000-0000-0000-0000-000000000006', 'pls', 'system', 'cashflow'),
    ('00000000-0000-0000-0000-000000000007', 'pls', 'system', 'calypso'),
    ('00000000-0000-0000-0000-000000000008', 'pls', 'create-role', null),
    ('00000000-0000-0000-0000-000000000009', 'pls', 'system', '*'),
    ('00000000-0000-0000-0000-000000000010', 'pls', 'create-role', null),
    ('00000000-0000-0000-0000-000000000011', 'pls', 'role', '*');

insert into roles_permissions (permission_instance_id, role_id) values
    ('00000000-0000-0000-0000-000000000000', 'dfunkt'),
    ('00000000-0000-0000-0000-000000000001', 'dfunkt'),
    ('00000000-0000-0000-0000-000000000002', 'dfunkt'),
    ('00000000-0000-0000-0000-000000000003', 'kassor'),
    ('00000000-0000-0000-0000-000000000004', 'kassor'),
    ('00000000-0000-0000-0000-000000000005', 'drek'),
    ('00000000-0000-0000-0000-000000000006', 'kassor'),
    ('00000000-0000-0000-0000-000000000007', 'komm'),
    ('00000000-0000-0000-0000-000000000008', 'drek'),
    ('00000000-0000-0000-0000-000000000009', 'dsys'),
    ('00000000-0000-0000-0000-000000000010', 'dsys'),
    ('00000000-0000-0000-0000-000000000011', 'dsys');

commit;
