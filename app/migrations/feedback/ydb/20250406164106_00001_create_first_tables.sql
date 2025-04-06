-- +goose Up
-- +goose StatementBegin
CREATE TABLE `feedback/reviews` (
    id String NOT NULL,
    product_id String NOT NULL,
    user_id String NOT NULL,
    rating Double NOT NULL,
    review Utf8 NOT NULL,
    created_at Datetime NOT NULL,
    updated_at Datetime NOT NULL,
    PRIMARY KEY (id),
    INDEX idx_product_created_at GLOBAL ASYNC ON (product_id, created_at)
);
CREATE TABLE `feedback/purchases` (
    user_id String NOT NULL,
    product_id String NOT NULL,
    order_id String NOT NULL,
    created_at Datetime NOT NULL,
    PRIMARY KEY (user_id, product_id, order_id)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `feedback/reviews`;
DROP TABLE `feedback/purchases`;
-- +goose StatementEnd
