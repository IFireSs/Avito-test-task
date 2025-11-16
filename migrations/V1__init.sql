CREATE TABLE teams
(
    team_name  TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


CREATE TABLE users
(
    user_id   TEXT PRIMARY KEY,
    username  TEXT    NOT NULL,
    team_name TEXT    NOT NULL REFERENCES teams (team_name) ON DELETE RESTRICT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_users_team ON users (team_name);
CREATE INDEX idx_users_team_active ON users (team_name, is_active);

CREATE TABLE pull_requests
(
    pull_request_id   TEXT PRIMARY KEY,
    pull_request_name TEXT        NOT NULL,
    author_id         TEXT        NOT NULL REFERENCES users (user_id) ON DELETE RESTRICT,
    status            TEXT        NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    merged_at         TIMESTAMPTZ
);

CREATE INDEX idx_pr_author ON pull_requests (author_id);

CREATE TABLE pull_request_reviewers
(
    pull_request_id TEXT NOT NULL REFERENCES pull_requests (pull_request_id) ON DELETE CASCADE,
    reviewer_id     TEXT NOT NULL REFERENCES users (user_id) ON DELETE RESTRICT,
    PRIMARY KEY (pull_request_id, reviewer_id)
);

CREATE INDEX idx_pr_reviewers_reviewer ON pull_request_reviewers (reviewer_id);