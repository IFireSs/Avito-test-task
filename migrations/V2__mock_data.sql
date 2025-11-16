INSERT INTO teams (team_name)
VALUES ('backend'),
       ('data'),
       ('frontend') ON CONFLICT (team_name) DO NOTHING;

INSERT INTO users (user_id, username, team_name, is_active)
VALUES ('u_backend_lead', 'backend_lead', 'backend', TRUE),
       ('u_backend_dev1', 'backend_dev1', 'backend', TRUE),
       ('u_backend_dev2', 'backend_dev2', 'backend', TRUE),
       ('u_backend_dev3', 'backend_dev3', 'backend', TRUE),
       ('u_backend_intern', 'backend_intern', 'backend', FALSE),

       ('u_data_lead', 'data_lead', 'data', TRUE),
       ('u_data_dev1', 'data_dev1', 'data', TRUE),
       ('u_data_dev2', 'data_dev2', 'data', TRUE),
       ('u_data_single', 'data_single', 'data', TRUE),

       ('u_frontend_lead', 'frontend_lead', 'frontend', TRUE),
       ('u_frontend_dev1', 'frontend_dev1', 'frontend', TRUE),
       ('u_frontend_dev2', 'frontend_dev2', 'frontend', TRUE),
       ('u_frontend_int', 'frontend_int', 'frontend', FALSE) ON CONFLICT (user_id) DO NOTHING;

INSERT INTO pull_requests (pull_request_id,
                           pull_request_name,
                           author_id,
                           status,
                           created_at,
                           merged_at)
VALUES ('pr_backend_1',
        'backend: implement feature X',
        'u_backend_lead',
        'OPEN',
        now() - interval '1 day',
        NULL),
       ('pr_backend_2',
        'backend: fix bug Y',
        'u_backend_dev1',
        'MERGED',
        now() - interval '3 days',
        now() - interval '2 days'),
       ('pr_backend_3',
        'backend: refactor module Z',
        'u_backend_dev2',
        'OPEN',
        now() - interval '5 hours',
        NULL),
       ('pr_backend_4',
        'backend: cleanup legacy',
        'u_backend_dev3',
        'OPEN',
        now() - interval '10 hours',
        NULL) ON CONFLICT (pull_request_id) DO NOTHING;

INSERT INTO pull_requests (pull_request_id,
                           pull_request_name,
                           author_id,
                           status,
                           created_at,
                           merged_at)
VALUES ('pr_data_1',
        'data: prepare dataset Z',
        'u_data_lead',
        'OPEN',
        now() - interval '2 days',
        NULL),
       ('pr_data_2',
        'data: build reporting pipeline',
        'u_data_dev1',
        'OPEN',
        now() - interval '12 hours',
        NULL),
       ('pr_data_3',
        'data: optimize query',
        'u_data_dev2',
        'MERGED',
        now() - interval '4 days',
        now() - interval '1 day') ON CONFLICT (pull_request_id) DO NOTHING;

INSERT INTO pull_requests (pull_request_id,
                           pull_request_name,
                           author_id,
                           status,
                           created_at,
                           merged_at)
VALUES ('pr_frontend_1',
        'frontend: add new landing',
        'u_frontend_lead',
        'OPEN',
        now() - interval '6 hours',
        NULL),
       ('pr_frontend_2',
        'frontend: fix layout',
        'u_frontend_dev1',
        'OPEN',
        now() - interval '1 day',
        NULL),
       ('pr_frontend_3',
        'frontend: remove dead code',
        'u_frontend_dev2',
        'MERGED',
        now() - interval '3 days',
        now() - interval '2 days') ON CONFLICT (pull_request_id) DO NOTHING;

INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
VALUES ('pr_backend_1', 'u_backend_dev1'),
       ('pr_backend_1', 'u_backend_dev2'),

       ('pr_backend_2', 'u_backend_dev2'),
       ('pr_backend_3', 'u_backend_intern'),

       ('pr_backend_4', 'u_backend_dev2'),
       ('pr_backend_4', 'u_backend_dev3') ON CONFLICT DO NOTHING;

INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
VALUES ('pr_data_1', 'u_data_dev1'),

       ('pr_data_2', 'u_data_single'),

       ('pr_data_3', 'u_data_dev2') ON CONFLICT DO NOTHING;

INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
VALUES ('pr_frontend_1', 'u_frontend_dev1'),
       ('pr_frontend_1', 'u_frontend_dev2'),
       ('pr_frontend_2', 'u_frontend_dev2') ON CONFLICT DO NOTHING;