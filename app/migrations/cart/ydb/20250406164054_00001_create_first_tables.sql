-- +goose Up
-- +goose StatementBegin
CREATE TABLE `cart/positions` (
    user_id Utf8 NOT NULL,
    product_id Utf8 NOT NULL,
    count Uint32 NOT NULL,
    PRIMARY KEY (user_id, product_id),
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `cart/positions`;
-- +goose StatementEnd
