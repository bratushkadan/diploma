-- +goose Up
-- +goose StatementBegin
CREATE TABLE `products/products` (
    id Uuid,
    seller_id Utf8,
    name Utf8 NOT NULL,
    description Utf8,
    picture_urls List<String> NOT NULL,
    stock Uint32 NOT NULL,
    created_at Datetime NOT NULL,
    updated_at Datetime,
    PRIMARY KEY (id),
    INDEX idx_seller_id on(seller_id),
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `products/products`;
-- +goose StatementEnd
