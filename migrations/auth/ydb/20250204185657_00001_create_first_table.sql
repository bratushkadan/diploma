-- +goose Up
-- +goose StatementBegin
CREATE TABLE accounts (
    id BIGSERIAL,
    name Utf8 NOT NULL,
    password Utf8 NOT NULL,
    email Utf8 NOT NULL,
    type String NOT NULL,
    created_at Timestamp NOT NULL,
    activated_at Datetime,
    PRIMARY KEY (id),
    INDEX idx_email_uniq GLOBAL UNIQUE SYNC on(email),
);
CREATE TABLE refresh_tokens (
    id BIGSERIAL,
    user_id Int64 NOT NULL,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    PRIMARY KEY (id, user_id),
) WITH (
    TTL = Interval("P30D") ON expires_at
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE accounts;
DROP TABLE refresh_tokens;
-- +goose StatementEnd
