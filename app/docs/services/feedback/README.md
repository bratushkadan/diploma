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
    created_at Datetime NOT NULL,
    updated_at Datetime NOT NULL,
    PRIMARY KEY (id),
    INDEX idx_product_created_at GLOBAL ASYNC ON (product_id, created_at)
);
```

```sql
CREATE TABLE `feedback/purchases` (
    user_id String NOT NULL,
    product_id String NOT NULL,
    order_id String NOT NULL,
    created_at Datetime NOT NULL,
    PRIMARY KEY (user_id, product_id, order_id)
)
```

## SEED(S) use cases

- Leave feedback on a purchased product

## Private endpoints

- Process published message on contents of `completed` order.

## Run

### Setup env and run

## CURLs for testing
