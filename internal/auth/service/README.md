# Services

Services are entities that embed the dependencies (secondary adapters).

## DTOs

Methods of a service receive an object parameter that is called DTO (Data Transfer Object).
This way driver adapters are not concerned with the domain layer types, and driver adapters are only concerned with the types defined in the service layer.

An example on DTOs:
```go
type CreateAccountReq struct {
    Username string
}

type CreateAccountAResp struct {
    UserID string
}
```
