# Floral

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

