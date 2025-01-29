# cmd/ directory arrangement

In hexagonal architecture, one could organize the cmd/ directory (for a single service repository, this service is "auth" for current directory)
1. `cmd/main.go`: driver adapters (REST, gRPC, CLI) are swappable via configuration parameters (implemented by factory methods);
2. `cmd/{rest,gprc,cli}/main.go`: driver adapters are separated via directories.
