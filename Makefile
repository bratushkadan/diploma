.PHONY: migrate_auth_create_ydb
migrate_auth_create_ydb:
	@sh -c "if [ -z "$$MIGRATION_NAME" ]; then echo 'Error: provide the \"MIGRATION_NAME\" env variable like MIGRATION_NAME=\"00001_create_first_table\"' >&2 && exit 1; else :; fi" && \
		echo scripts/migrate create "$${MIGRATION_NAME}" sql
.PHONY: migrate_auth_DANGER_DOWN_ydb
migrate_auth_DANGER_DOWN_ydb:
	@scripts/migrate down
.PHONY: migrate_auth_up_ydb
migrate_auth_up_ydb:
	@scripts/migrate up
.PHONY: migrate_auth_up_by_one_ydb
migrate_auth_up_by_one_ydb:
	@scripts/migrate up-by-one

.PHONY: migrate_auth_create_pg
migrate_auth_create_pg:
	@migrate create -ext=sql -dir "./migrations/auth/postgres" -seq init
.PHONY: migrate_auth_up_pg
migrate_auth_up_pg:
	@migrate -path "./migrations/auth/postgres" -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" -verbose up
.PHONY: migrate_auth_down_pg
migrate_auth_down_pg:
	@migrate -path "./migrations/auth/postgres" -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" -verbose down 1

