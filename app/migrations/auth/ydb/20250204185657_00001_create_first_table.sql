-- +goose Up
-- +goose StatementBegin
CREATE TABLE accounts (
    id BigSerial,
    name Utf8 NOT NULL,
    password Utf8 NOT NULL,
    email Utf8 NOT NULL,
    type String NOT NULL,
    created_at Datetime NOT NULL,
    activated_at Datetime,
    PRIMARY KEY (id),
    INDEX idx_email_uniq GLOBAL UNIQUE SYNC on(email),
);
CREATE TABLE refresh_tokens (
    id BigSerial,
    account_id Utf8 NOT NULL,
    created_at Datetime NOT NULL,
    expires_at Datetime NOT NULL,
    PRIMARY KEY (id),
    INDEX idx_account_id GLOBAL SYNC on(account_id)
) WITH (
    TTL = Interval("P30D") ON expires_at
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE accounts;
DROP TABLE refresh_tokens;
-- +goose StatementEnd
