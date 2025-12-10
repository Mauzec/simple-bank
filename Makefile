.PHONY: postgres
postgres:
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres

.PHONY: createdb
createdb:
	docker exec -it postgres createdb --username=root --owner=root simple_bank

.PHONY: dropdb
dropdb:
	docker exec -it postgres dropdb simple_bank

.PHONY: migrateup
migrateup: 
	migrate -path db/migrate -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

.PHONY: migratedown
migratedown:
	migrate -path db/migrate -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

.PHONY: migrateup1
migrateup1: 
	migrate -path db/migrate -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

.PHONY: migratedown1
migratedown1:
	migrate -path db/migrate -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

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
	mockgen -package mockdb -destination db/mock/store.go github.com/mauzec/simple-bank/db/sqlc Store