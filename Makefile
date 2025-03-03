.PHONY: postgres
postgres:
	docker run --name postgres -p 5431:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres

.PHONY: createdb
createdb:
	docker exec -it postgres createdb --username=root --owner=root simple_bank

.PHONY: dropdb
dropdb:
	docker exece -it postgres dropdb simple_bank

.PHONY: migrateup
migrateup: 
	migrate -path db/migrate -database "postgresql://root:secret@localhost:5431/simple_bank?sslmode=disable" -verbose up

.PHONY: migratedown
migratedown:
	migrate -path db/migrate -database "postgresql://root:secret@localhost:5431/simple_bank?sslmode=disable" -verbose down

.PHONY: sqlc
sqlc:
	sqlc generate

.PHONY: test
test:
	@if command -v gotestsum > /dev/null; then \
		gotestsum --format testname; \
	else \
		go test ./...; \
	fi
