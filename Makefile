include .env
MIGRATIONS_PATH = ./cmd/migrate/migrations

.PHONY: migrate-create
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -database $(DB_ADDR) -path $(MIGRATIONS_PATH) up

.PHONY: migrate-up-force
migrate-up-force:
	@migrate -database $(DB_ADDR) -path $(MIGRATIONS_PATH) force $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-down
migrate-down:
	@migrate -database $(DB_ADDR) -path $(MIGRATIONS_PATH) down $(filter-out $@,$(MAKECMDGOALS))

.PHONY: seed
seed:
	@go run cmd/migrate/seed/main.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt