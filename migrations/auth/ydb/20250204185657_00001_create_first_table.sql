-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id SERIAL,
    name Utf8 NOT NULL,
    password Utf8 NOT NULL,
    email Utf8 NOT NULL,
    type String NOT NULL,
    created_at Timestamp,
    INDEX idx_email_uniq GLOBAL UNIQUE on(email),
    PRIMARY KEY (id),
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
