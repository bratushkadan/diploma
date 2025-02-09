# Database migrations

Database migrations are handled via [golang-migrate](https://github.com/golang-migrate/migrate).

## Commands

To create a migration (e.g. for a service *auth*), run the following command:

```bash
make migrate_auth_create
```

To migrate, run the following command:
```bash
POSTGRES_USER=root POSTGRES_PASSWORD="" POSTGRES_HOST=localhost POSTGRES_PORT=5432 POSTGRES_DB=auth make migrate_auth_up
```

