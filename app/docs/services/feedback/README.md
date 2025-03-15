# Feedback microservice

## Data Model

[DB Diagram](https://dbdiagram.io/d/ecom-67b96d09263d6cf9a01083b2)

YDB Schema:

```sql
CREATE TABLE `feedback/reviews` (
    id String NOT NULL,
    product_id String NOT NULL,
    author_id String NOT NULL,
    rating Double NOT NULL,
    review Utf8 NOT NULL,
    INDEX idx_product_id GLOBAL ASYNC ON (product_id)
);
```

## Run

### Setup env and run

