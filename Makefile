.PHONY: gen
gen: go-generate	

.PHONY: go-gen
go-gen:
	@mkdir -p pb
	@go generate ./...

.PHONY: migrate_auth_create
migrate_auth_create:
	@migrate create -ext=sql -dir "./migrations/auth" -seq init


.PHONY: migrate_auth_up
migrate_auth_up:
	@migrate -path "./migrations/auth" -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" -verbose up

.PHONY: migrate_auth_down
migrate_auth_down:
	@migrate -path "./migrations/auth" -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" -verbose down 1
