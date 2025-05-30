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

.PHONY: migrate_cart_create
migrate_cart_create:
	@export SERVICE=cart && \
		sh -c "if [ -z "$$MIGRATION_NAME" ]; then echo 'Error: provide the \"MIGRATION_NAME\" env variable like MIGRATION_NAME=\"00001_create_first_table\"' >&2 && exit 1; else :; fi" && \
		scripts/migrate create "$${MIGRATION_NAME}" sql
.PHONY: migrate_cart_DANGER_DOWN
migrate_cart_DANGER_DOWN:
	@export SERVICE=cart && \
		scripts/migrate down
.PHONY: migrate_cart_up
migrate_cart_up:
	@export SERVICE=cart && \
		scripts/migrate up
.PHONY: migrate_cart_up_by_one
migrate_cart_up_by_one:
	@export SERVICE=cart && \ scripts/migrate up-by-one

.PHONY: migrate_orders_create
migrate_orders_create:
	@export SERVICE=orders && \
		sh -c "if [ -z "$$MIGRATION_NAME" ]; then echo 'Error: provide the \"MIGRATION_NAME\" env variable like MIGRATION_NAME=\"00001_create_first_table\"' >&2 && exit 1; else :; fi" && \
		scripts/migrate create "$${MIGRATION_NAME}" sql
.PHONY: migrate_orders_DANGER_DOWN
migrate_orders_DANGER_DOWN:
	@export SERVICE=orders && \
		scripts/migrate down
.PHONY: migrate_orders_up
migrate_orders_up:
	@export SERVICE=orders && \
		scripts/migrate up
.PHONY: migrate_orders_up_by_one
migrate_orders_up_by_one:
	@export SERVICE=orders && \ scripts/migrate up-by-one

.PHONY: migrate_feedback_create
migrate_feedback_create:
	@export SERVICE=feedback && \
		sh -c "if [ -z "$$MIGRATION_NAME" ]; then echo 'Error: provide the \"MIGRATION_NAME\" env variable like MIGRATION_NAME=\"00001_create_first_table\"' >&2 && exit 1; else :; fi" && \
		scripts/migrate create "$${MIGRATION_NAME}" sql
.PHONY: migrate_feedback_DANGER_DOWN
migrate_feedback_DANGER_DOWN:
	@export SERVICE=feedback && \
		scripts/migrate down
.PHONY: migrate_feedback_up
migrate_feedback_up:
	@export SERVICE=feedback && \
		scripts/migrate up
.PHONY: migrate_feedback_up_by_one
migrate_feedback_up_by_one:
	@export SERVICE=feedback && \ scripts/migrate up-by-one

# API Generation requires Terraform to be applied first
# ../terraform/tf apply -target local_file.public_private_api_oapi_spec - generates file "public_private_api_oapi_spec.yaml" to use spec from.

.PHONY: generate_api
generate_api: generate_api_products generate_api_catalog generate_api_cart generate_api_orders generate_api_feedback

.PHONY: generate_api_products
generate_api_products: public_private_api_oapi_spec.yaml
	@mkdir -p ./internal/products/presentation/generated
	@cat public_private_api_oapi_spec.yaml | yq -M 'del(.servers)' > ./internal/products/presentation/oapi/api.yaml
	@go generate ./internal/products/...
.PHONY: generate_api_catalog
generate_api_catalog:
	@mkdir -p ./internal/catalog/presentation/generated
	@cat public_private_api_oapi_spec.yaml | yq -M 'del(.servers)' > ./internal/catalog/presentation/oapi/api.yaml
	@go generate ./internal/catalog/...
.PHONY: generate_api_cart
generate_api_cart:
	@mkdir -p ./internal/cart/presentation/generated
	@cat public_private_api_oapi_spec.yaml | yq -M 'del(.servers)' > ./internal/cart/presentation/oapi/api.yaml
	@go generate ./internal/cart/...
.PHONY: generate_api_orders
generate_api_orders:
	@mkdir -p ./internal/orders/presentation/generated
	@cat public_private_api_oapi_spec.yaml | yq -M 'del(.servers)' > ./internal/orders/presentation/oapi/api.yaml
	@go generate ./internal/orders/...
.PHONY: generate_api_feedback
generate_api_feedback:
	@mkdir -p ./internal/feedback/presentation/generated
	@cat public_private_api_oapi_spec.yaml | yq -M 'del(.servers)' > ./internal/feedback/presentation/oapi/api.yaml
	@go generate ./internal/feedback/...

.PHONY: migrate_auth_create_pg
migrate_auth_create_pg:
	@migrate create -ext=sql -dir "./migrations/auth/postgres" -seq init
.PHONY: migrate_auth_up_pg
migrate_auth_up_pg:
	@migrate -path "./migrations/auth/postgres" -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" -verbose up
.PHONY: migrate_auth_down_pg
migrate_auth_down_pg:
	@migrate -path "./migrations/auth/postgres" -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" -verbose down 1

