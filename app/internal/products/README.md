# Products service

Layered Architecture (3-L):

*http* is Presentation layer
*app* is Business Logic Layer
*store* is Persistence Layer

## Get OpenAPI spec for services to enable code generation

```sh
yc serverless api-gateway get-spec auth-service-api-gw
```
