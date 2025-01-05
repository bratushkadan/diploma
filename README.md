# Floral

## Run

### Auth service

```bash
go run github.com/bratushkadan/floral/cmd/auth
```

#### Or this way

```bash
AUTH_JWT_PRIVATE_KEY_PATH=./pkg/auth/test_fixtures/private.key AUTH_JWT_PUBLIC_KEY_PATH=./pkg/auth/test_fixtures/public.key go run ./cmd/auth/auth.go
```

### Test Auth service is running

```bash
grpcurl -plaintext -d '{"id": 1}' '127.0.0.1:48612' floral.auth.v1.UserService/GetUser
```

## gRPC & Go dependencies

Install:
1. https://grpc.io/docs/protoc-installation/
2. https://protobuf.dev/reference/go/go-generated/
3. https://grpc.io/docs/languages/go/quickstart/

## Adding/updating proto types/gRPC services

1. Write new proto code in ./proto;
2. Update the `./pb/generate.go` file;
3. Run `make go-gen`; 
4. Import pb in code (example: `github.com/bratushkadan/floral/pb/floral/auth/v1`).

## HTTP API Docs

### Auth

#### Error Response

Sample:
```json
{
  "errors": [
    {
      "code": 1,
      "message": "bad request"
    }
  ]
}
```

#### Endpoints

##### `POST /api/v1/users/:register`

Request sample:
```json
{
  "name": "danila",
  "email": "foobar@yahoo.com",
  "password": "secretpass123"
}
```

Response sample:
```json
{
  "id": "i1qwk6jcuwjeqzy",
  "name": "danila"
}
```


##### `POST /api/v1/users/:authenticate`

Request sample:
```json
{
  "email": "foobar@yahoo.com",
  "password": "secretpass123"
}
```

Response sample:
```json
{
  "refresh_token": "..."
}
```

##### `POST /api/v1/users/:renewRefreshToken`

Request sample:
```json
{
  "refresh_token": "..."
}
```

Response sample:
```json
{
  "refresh_token": "..."
}
```
