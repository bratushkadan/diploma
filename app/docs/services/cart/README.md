# Cart microservice

## Data Model

[DB Diagram](https://dbdiagram.io/d/ecom-67b96d09263d6cf9a01083b2)

YDB Schema:

```sql
CREATE TABLE `cart/positions` (
    user_id Utf8 NOT NULL,
    product_id Utf8 NOT NULL,
    count Uint32 NOT NULL,
    PRIMARY KEY (user_id, product_id),
);
```

## SEED(s) use cases

- Add product to cart (or change count of products in cart)
- Delete product from cart

## Private endpoints

- Process publish cart contents (process event/message)
- Process clear cart (process event/message)

## Details

No more than 25 distinct items in cart.

If a user has products from one seller in their cart and a product from another seller is added to the cart, cart is first cleared and then product from another seller is added.

## Run

### Setup env and run

## CURLs for testing
