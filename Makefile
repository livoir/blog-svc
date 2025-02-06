PG_HOST ?= localhost
PG_PORT ?= 5432
PG_USER ?= postgres
PG_PASSWORD ?= postgres
PG_DATABASE ?= blog
PRIVATE_KEY ?= server.key
PUBLIC_KEY ?= server.pem

connection := "host=$(PG_HOST) port=$(PG_PORT) user=$(PG_USER) password=$(PG_PASSWORD) dbname=$(PG_DATABASE) sslmode=$(PG_SSL_MODE)"
dir := ./migrations
goose := goose -dir $(dir) postgres $(connection)

migration-status:
	$(goose) status
migration-create:
	$(goose) create $(name) sql
migration-up:
	$(goose) up

test:
	go test -v -cover ./...

generate-cert:
	openssl genrsa -out configs/$(PRIVATE_KEY) 2048
	openssl rsa -in configs/$(PRIVATE_KEY) -pubout -out configs/$(PUBLIC_KEY)