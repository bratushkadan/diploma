-- +goose Up
-- +goose StatementBegin
CREATE TABLE `orders/orders` (
  id Utf8 NOT NULL,
  user_id Utf8 NOT NULL,
  -- For ease of designing and implementing business processes only online payments are allowed
  -- online_payment Bool NOT NULL DEFAULT false,
  status Utf8 NOT NULL,
  created_at Datetime NOT NULL,
  updated_at Datetime NOT NULL,
  PRIMARY KEY (id),
  INDEX idx_list_orders GLOBAL ASYNC on (user_id, created_at)
);
CREATE TABLE `orders/order_items` (
  order_id Utf8 NOT NULL,
  product_id Utf8 NOT NULL,
  name Utf8 NOT NULL,
  seller_id Utf8 NOT NULL,
  count Uint32 NOT NULL,
  price Double NOT NULL,
  picture Utf8,
  PRIMARY KEY (order_id, product_id)
);
CREATE TABLE `orders/payments` (
  id Utf8 NOT NULL,
  order_id Utf8 NOT NULL,
  amount Double NOT NULL,
  currency_iso_4217 Uint32 NOT NULL,
  provider Json NOT NULL,
  created_at Timestamp NOT NULL,
  updated_at Timestamp NOT NULL,
  refunded_at Timestamp,
  PRIMARY KEY (id),
  INDEX idx_order_id GLOBAL ASYNC on (order_id)
);
CREATE TABLE `orders/operations` (
  id Utf8 NOT NULL,
  type Utf8 NOT NULL,
  status Utf8 NOT NULL,
  details Utf8,
  user_id Utf8 NOT NULL,
  order_id Utf8,
  created_at Timestamp NOT NULL,
  updated_at Timestamp NOT NULL,
  PRIMARY KEY (id),
  INDEX idx_status GLOBAL ASYNC on (status),
  INDEX idx_order_id GLOBAL ASYNC on (order_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `orders/orders`;
DROP TABLE `orders/order_items`;
DROP TABLE `orders/payments`;
DROP TABLE `orders/operations`;
-- +goose StatementEnd
