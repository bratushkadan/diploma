# Order microservice

## Data Model

[DB Diagram](https://dbdiagram.io/d/ecom-67b96d09263d6cf9a01083b2)

YDB Schema:

```sql
CREATE TABLE `orders/orders` (
  order_id String NOT NULL,
  user_id String NOT NULL,
  -- contents of order
  contents Json NOT NULL,
  -- For ease of designing and implementing business processes only online payments are allowed
  -- online_payment Bool NOT NULL DEFAULT false,
  status String NOT NULL,
  PRIMARY KEY order_id,
  INDEX idx_user_id GLOBAL ASYNC on (user_id)
);
```

`contents` field schema:
```go
type OrderContentsItem struct {
    ProductId string `json:"product_id"`
    Name string `json:"name"`
    SellerId string `json:"seller_id"`
    Count int32 `json:"count"`
    Price float64 `json:"price"`
    // Picture url
    Picture *string `json:"picture"`
}
type OrderContents = OrderContentsItem[] 
```

```sql
CREATE TABLE `orders/payments` (
  id String NOT NULL,
  order_id String NOT NULL,
  provider Json NOT NULL,
  created_at Timestamp NOT NULL,
  updated_at Timestamp NOT NULL,
  refunded_at Timestamp,
  PRIMARY KEY id,
  INDEX idx_order_id GLOBAL ASYNC on (order_id)
);
```

## Private endpoints

- Process cart contents ("cart contents" event/message)
- Process reserved products contents ("reserved products contents" event/message)
- Cancel unpaid orders (invoked by Serverless Trigger)

## General idea

Orders that are older than one hour and are not paid online (if not paid by cash) are cancelled. Order cancellation is scheduled regularly.

## Run

### Setup env and run

## CURLs for testing
