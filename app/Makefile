.PHONY: auth_run_integration_tests
auth_run_integration_tests:
	@YDB_SERVICE_ACCOUNT_KEY_FILE_CREDENTIALS="$$(scripts/ydb_access_token.sh)" \
		YDB_ENDPOINT="$$(../terraform/tf output -json | jq -cMr .ydb.value.full_endpoint)" \
		YDB_AUTH_METHOD=environ \
		go run cmd/auth/integration_tests/main.go

.PHONY: migrate_auth_create_ydb
migrate_auth_create_ydb:
	@export SERVICE=auth && \
		sh -c "if [ -z "$$MIGRATION_NAME" ]; then echo 'Error: provide the \"MIGRATION_NAME\" env variable like MIGRATION_NAME=\"00001_create_first_table\"' >&2 && exit 1; else :; fi" && \
		scripts/migrate create "$${MIGRATION_NAME}" sql
.PHONY: migrate_auth_DANGER_DOWN_ydb
migrate_auth_DANGER_DOWN_ydb:
	@export SERVICE=auth && \
		scripts/migrate down
.PHONY: migrate_auth_up_ydb
migrate_auth_up_ydb:
	@export SERVICE=auth && \
		scripts/migrate up
.PHONY: migrate_auth_up_by_one_ydb
migrate_auth_up_by_one_ydb:
	@export SERVICE=auth && \
		scripts/migrate up-by-one

.PHONY: migrate_products_create
migrate_products_create:
	@export SERVICE=products && \
		sh -c "if [ -z "$$MIGRATION_NAME" ]; then echo 'Error: provide the \"MIGRATION_NAME\" env variable like MIGRATION_NAME=\"00001_create_first_table\"' >&2 && exit 1; else :; fi" && \
		scripts/migrate create "$${MIGRATION_NAME}" sql
.PHONY: migrate_products_DANGER_DOWN
migrate_products_DANGER_DOWN:
	@export SERVICE=products && \
		scripts/migrate down
.PHONY: migrate_products_up
migrate_products_up:
	@export SERVICE=products && \
		scripts/migrate up
.PHONY: migrate_products_up_by_one
migrate_products_up_by_one:
	@export SERVICE=products && \ scripts/migrate up-by-one

.PHONY: generate_api_products
generate_api_products:
	@yc serverless api-gateway get-spec auth-service-api-gw | yq -M 'del(.servers)' > ./internal/products/presentation/oapi/api.yaml && \
		go generate ./internal/products/...

.PHONY: migrate_auth_create_pg
migrate_auth_create_pg:
	@migrate create -ext=sql -dir "./migrations/auth/postgres" -seq init
.PHONY: migrate_auth_up_pg
migrate_auth_up_pg:
	@migrate -path "./migrations/auth/postgres" -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" -verbose up
.PHONY: migrate_auth_down_pg
migrate_auth_down_pg:
	@migrate -path "./migrations/auth/postgres" -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" -verbose down 1

