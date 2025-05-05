ifeq (,$(wildcard ./config/app.env))
  $(error "Create a app.env file based on app.env.example")
endif
include ./config/app.env
export

DB_SOURCE := $(DB_SOURCE)
POSTGRES_USER := $(POSTGRES_USER)
POSTGRES_PASSWORD := $(POSTGRES_PASSWORD)
DB_PORT := $(DB_PORT)
DB_NAME := $(DB_NAME)
SSL_ENABLE := $(SSL_ENABLE)

.PHONE:
network:
	docker network create bank-network

.PHONY: postgres
postgres:
	docker run \
	--name postgres --network bank-network \
	-p $(DB_PORT):$(DB_PORT) -e POSTGRES_USER=$(POSTGRES_USER) \
	-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) -d postgres

.PHONY: createdb
createdb:
	docker exec -it postgres createdb \
	--username=$(POSTGRES_USER) --owner=$(POSTGRES_USER) $(DB_NAME)

.PHONY: dropdb
dropdb:
	docker exec -it postgres dropdb $(DB_NAME)

.PHONY: migrateup
migrateup: 
	migrate -path db/migrate \
	-database postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)$(SSL_ENABLE) \
	-verbose up

.PHONY: migratedown
migratedown:
	migrate -path db/migrate \
	-database postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)$(SSL_ENABLE) \
	-verbose down

.PHONY: migrateup1
migrateup1: 
	migrate -path db/migrate \
	-database postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)$(SSL_ENABLE) \
	-verbose up 1

.PHONY: migratedown1
migratedown1:
	migrate -path db/migrate \
	-database postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)$(SSL_ENABLE) \
	-verbose down 1

.PHONY: full_restart
full_restart:
	@read -p "Are you sure? This will delete all data. Type 'yes' to continue: " confirm && [ "$$confirm" = "yes" ] && \
	$(MAKE) dropdb && \
	$(MAKE) createdb && \
	$(MAKE) migrateup || \
	echo "Operation cancelled."

.PHONY: sqlc
sqlc:
	sqlc generate

.PHONY: test
test:
	@if command -v gotestsum > /dev/null; then \
		gotestsum --debug --format testname; \
	else \
		go test ./...; \
	fi

.PHONY: testv
testv:
	@if command -v gotestsum > /dev/null; then \
		gotestsum --debug --format standard-verbose; \
	else \
		go test -v ./...; \
	fi

.PHONY: server
server:
	go run cmd/server/main.go

.PHONY: mockdb
mockdb:
	mockgen -package mockdb \
	-destination db/mock/store.go github.com/mauzec/simple-bank/db/sqlc Store