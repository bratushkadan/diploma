# E-com platform

## Architecture

- [Project Architectural Decision Records](./app/docs/adr)
- [Architecture Diagrams (drawio)](./app/docs/architecture)

## [Services](./app/docs/services)

### [Auth](./app/docs/services/auth)
### [Products](./app/docs/services/products)
### [Catalog](./app/docs/services/catalog)

## Setup

### [Terraform](./terraform)

### Per-service setup

YDB:
- [document API for Auth](./app/docs/services/auth)
- [regular cluster for almost all other services & CDC](./app/Makefile)

## Email Sending

1. https://id.yandex.ru/security/app-passwords - add password

[Official Yandex Mail docs](https://yandex.ru/support/yandex-360/business/mail/ru/web/security/oauth)
