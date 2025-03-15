# Cart microservice

## Data Model

[DB Diagram](https://dbdiagram.io/d/ecom-67b96d09263d6cf9a01083b2)

YDB Schema:

```sql
CREATE TABLE `cart/cart` (
    user_id Utf8 NOT NULL,
    product_id Utf8 NOT NULL,
    count Uint32 NOT NULL,
    -- name Utf8 NOT NULL,
    -- pictures Json NOT NULL,
    -- price Double NOT NULL,
    PRIMARY KEY (user_id, product_id),
);
```

## Details

If a user has products from one seller in their cart and a product from another seller is added to the cart, cart is first cleared and then product from another seller is added.

## Run

### Setup env and run
