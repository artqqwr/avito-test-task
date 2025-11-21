-- +goose Up
CREATE TABLE teams (
                       name VARCHAR(255) PRIMARY KEY
);

CREATE TABLE users (
                       id VARCHAR(255) PRIMARY KEY,
                       username VARCHAR(255) NOT NULL,
                       team_name VARCHAR(255) NOT NULL REFERENCES teams(name),
                       is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE pull_requests (
                               id VARCHAR(255) PRIMARY KEY,
                               name TEXT NOT NULL,
                               author_id VARCHAR(255) NOT NULL REFERENCES users(id),
                               status VARCHAR(50) NOT NULL DEFAULT 'OPEN',
                               created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                               merged_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE pr_reviewers (
                              pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
                              reviewer_id VARCHAR(255) NOT NULL REFERENCES users(id),
                              PRIMARY KEY (pull_request_id, reviewer_id)
);

CREATE INDEX idx_users_team ON users(team_name);
CREATE INDEX idx_users_active ON users(is_active);
CREATE INDEX idx_pr_author ON pull_requests(author_id);

-- +goose Down
DROP TABLE pr_reviewers;
DROP TABLE pull_requests;
DROP TABLE users;
DROP TABLE teams;