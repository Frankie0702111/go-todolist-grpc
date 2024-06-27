include app.env

DOCKER = docker compose exec server
YMD = _$$(date +'%Y%m%d')
number?=3
table :=
LOG_DIRS := internal/pkg/db/target internal/service/target

proto:
		protoc --proto_path=api/proto --go_out=api/pb --go_opt=paths=source_relative \
		--go-grpc_out=api/pb --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=api/pb --grpc-gateway_opt=paths=source_relative \
		api/proto/*.proto

grpcui:
		grpcui -plaintext ${APP_HOST}:${GRPC_SERVER_PORT}

# make migrate-create table=xxxx
migrate-create:
		$(DOCKER) migrate create -ext sql -dir ./internal/migrations -seq create_$(table)_table

# make migrate-test-up number=x
migrate-test-up:
		$(DOCKER) migrate -path ./internal/migrations -database "$(DB)://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/test_db?sslmode=disable" up $(number)

# make migrate-test-down number=x
migrate-test-down:
		$(DOCKER) migrate -path ./internal/migrations -database "$(DB)://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/test_db?sslmode=disable" down $(number)

clean-logs:
	@echo "Cleaning log directories..."
	@for dir in $(LOG_DIRS); do \
		if [ -d "$$dir" ]; then \
			echo "Removing $$dir"; \
			rm -rf "$$dir"; \
		else \
			echo "$$dir does not exist"; \
		fi \
	done

go-test:
	@set -e; \
	make migrate-test-up; \
	$(DOCKER) go test -v internal/config/config_test.go -json > ./target/log/config_test$(YMD).log; \
	$(DOCKER) go test -v internal/pkg/db/db_test.go -json > ./target/log/db_test$(YMD).log; \
	$(DOCKER) go test -v internal/pkg/hash/hash_test.go -json > ./target/log/hash_test$(YMD).log; \
	$(DOCKER) go test -v internal/pkg/util/jwt_test.go -json > ./target/log/jwt_test$(YMD).log; \
	$(DOCKER) go test -v internal/pkg/util/random_test.go -json > ./target/log/th_test$(YMD).log; \
	$(DOCKER) go test -v internal/pkg/util/th_test.go -json > ./target/log/th_test$(YMD).log; \
	$(DOCKER) go test -v internal/model/mod_user_test.go -json > ./target/log/th_test$(YMD).log; \
	$(DOCKER) go test -v internal/service/s_user_test.go -json > ./target/log/th_test$(YMD).log; \
	make migrate-test-down; \
	make clean-logs;

go-test-single:
	$(DOCKER) go test -v internal/pkg/util/random_test.go;

.PHONY: proto grpcui migrate-create migrate-test-up migrate-test-down clean-logs go-test go-test-single
