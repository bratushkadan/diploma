# Feedback microservice

## Data Model

[DB Diagram](https://dbdiagram.io/d/ecom-67b96d09263d6cf9a01083b2)

YDB Schema:

```sql
CREATE TABLE `feedback/reviews` (
    id String NOT NULL,
    product_id String NOT NULL,
    user_id String NOT NULL,
    rating Double NOT NULL,
    review Utf8 NOT NULL,
    INDEX idx_product_id GLOBAL ASYNC ON (product_id)
);
```

```sql
CREATE TABLE `feedback/purchases` (
    user_id String NOT NULL,
    product_id String NOT NULL,
    order_id String NOT NULL,
    purchased_at Datetime NOT NULL,
    PRIMARY KEY (buyer_id, product_id, order_id)
)
```

## SEED(S) use cases

- Leave feedback on a purchased product

## Private endpoints

- Process published message on contents of `completed` order.

## Run

### Setup env and run

## CURLs for testing
