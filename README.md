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

