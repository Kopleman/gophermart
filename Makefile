-include .env

MIGRATIONS_DIR   = ./sql/migrations/
GOLANGCI_LINT_CACHE?=/tmp/${APP_NAME}

.PHONY: build
build:
	go build -o ./cmd/gophermart/gophermart ./cmd/gophermart

.PHONY: run
run:
	go run ./cmd/gophermart/main.go

.PHONY: migrate
migrate:
	migrate -path "$(MIGRATIONS_DIR)" -database "$(DATABASE_URI)" $(filter-out $@,$(MAKECMDGOALS))

.PHONY: db-create-migration
db-create-migration:
	migrate create -ext sql -dir "$(MIGRATIONS_DIR)" $(filter-out $@,$(MAKECMDGOALS))

.PHONY: gensql
gensql:
	sqlc generate

genswagger:
	swag fmt -d ./cmd/gophermart
	swag fmt -d ./internal/controller
	swag init -o ./docs --dir ./internal/controller -g ../../cmd/gophermart/main.go --parseDependency

.PHONY: fix-field-alignment
fix-field-alignment:
	fieldalignment -fix ./...

.PHONY: golangci-lint-run
golangci-lint-run: _golangci-lint-rm-unformatted-report

.PHONY: _golangci-lint-reports-mkdir
_golangci-lint-reports-mkdir:
	mkdir -p ./golangci-lint

.PHONY: _golangci-lint-run
_golangci-lint-run: _golangci-lint-reports-mkdir
	-docker run --rm \
    -v $(shell pwd):/app \
    -v $(GOLANGCI_LINT_CACHE):/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.63.1 \
        golangci-lint run \
            -c .golangci.yml \
	> ./golangci-lint/report-unformatted.json

.PHONY: _golangci-lint-format-report
_golangci-lint-format-report: _golangci-lint-run
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json

.PHONY: _golangci-lint-rm-unformatted-report
_golangci-lint-rm-unformatted-report: _golangci-lint-format-report
	rm ./golangci-lint/report-unformatted.json

.PHONY: golangci-lint-clean
golangci-lint-clean:
	rm -rf ./golangci-lint

.PHONY: lint
lint: golangci-lint-clean golangci-lint-run