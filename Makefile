# TODO: добавить автоматическое инкреметнирование
.PHONY: migrate_auth_create_ydb
migrate_auth_create_ydb:
	@export YDB_API_ENDPOINT="$$(./terraform/tf output -json | jq -cMr .ydb.value.api_endpoint)" && \
		export YDB_DATABASE_PATH="$$(./terraform/tf output -json | jq -cMr .ydb.value.database_path)" && \
		export IAM_TOKEN="$$(yc iam create-token)" && \
		cd migrations/auth/ydb && \
		goose ydb "grpcs://$${YDB_API_ENDPOINT}$${YDB_DATABASE_PATH}?token=${IAM_TOKEN}&go_query_mode=scripting&go_fake_tx=scripting&go_query_bind=declare,numeric" \
		  create 00001_create_first_table sql
.PHONY: migrate_auth_create_pg
migrate_auth_create_pg:
	@migrate create -ext=sql -dir "./migrations/auth/postgres" -seq init


.PHONY: migrate_auth_up_pg
migrate_auth_up_pg:
	@migrate -path "./migrations/auth/postgres" -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" -verbose up

.PHONY: migrate_auth_down_pg
migrate_auth_down_pg:
	@migrate -path "./migrations/auth/postgres" -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" -verbose down 1

.PHONY: migrate_auth_up_ydb
migrate_auth_up_ydb:
	@export YDB_API_ENDPOINT="$$(./terraform/tf output -json | jq -cMr .ydb.value.api_endpoint)" && \
		export YDB_DATABASE_PATH="$$(./terraform/tf output -json | jq -cMr .ydb.value.database_path)" && \
		export IAM_TOKEN="$$(yc iam create-token)" && \
		cd migrations/auth/ydb && \
		goose ydb "grpcs://$${YDB_API_ENDPOINT}$${YDB_DATABASE_PATH}?token=${IAM_TOKEN}&go_query_mode=scripting&go_fake_tx=scripting&go_query_bind=declare,numeric" up
.PHONY: migrate_auth_up_by_one_ydb
migrate_auth_up_by_one_ydb:
	@export YDB_API_ENDPOINT="$$(./terraform/tf output -json | jq -cMr .ydb.value.api_endpoint)" && \
		export YDB_DATABASE_PATH="$$(./terraform/tf output -json | jq -cMr .ydb.value.database_path)" && \
		export IAM_TOKEN="$$(yc iam create-token)" && \
		cd migrations/auth/ydb && \
		goose ydb "grpcs://$${YDB_API_ENDPOINT}$${YDB_DATABASE_PATH}?token=${IAM_TOKEN}&go_query_mode=scripting&go_fake_tx=scripting&go_query_bind=declare,numeric" up-by-one
