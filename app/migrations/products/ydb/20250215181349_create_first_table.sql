-- +goose Up
-- +goose StatementBegin
CREATE TABLE `products/products` (
    id Uuid NOT NULL,
    seller_id Utf8 NOT NULL,
    name Utf8 NOT NULL,
    description Utf8,
    pictures Json NOT NULL,
    stock Uint32 NOT NULL,
    created_at Datetime NOT NULL,
    updated_at Datetime,
    PRIMARY KEY (id),
    INDEX idx_seller_id GLOBAL ASYNC ON (seller_id),
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `products/products`;
-- +goose StatementEnd
