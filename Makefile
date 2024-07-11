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

# make migrate-up number=x
migrate-up:
		$(DOCKER) migrate -path ./internal/migrations -database "$(DB)://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up $(number)

# make migrate-down number=x
migrate-down:
		$(DOCKER) migrate -path ./internal/migrations -database "$(DB)://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down $(number)

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
	start_time=$$(date +%s); \
	echo "Starting tests at $$(date)"; \
	make migrate-test-up; \
	go test -v internal/config/config_test.go -json > ./target/log/config_test$(YMD).log; \
	go test -v internal/pkg/db/db_test.go -json > ./target/log/db_test$(YMD).log; \
	go test -v internal/pkg/util/hash_test.go -json > ./target/log/hash_test$(YMD).log; \
	go test -v internal/pkg/util/jwt_test.go -json > ./target/log/jwt_test$(YMD).log; \
	go test -v internal/pkg/util/random_test.go -json > ./target/log/random_test$(YMD).log; \
	go test -v internal/pkg/util/th_test.go -json > ./target/log/th_test$(YMD).log; \
	go test -v internal/model/mod_user_test.go -json > ./target/log/mod_user_test$(YMD).log; \
	go test -v internal/service/s_user_test.go -json > ./target/log/s_user_test$(YMD).log; \
	go test -v internal/model/mod_category_test.go -json > ./target/log/mod_category_test$(YMD).log; \
	go test -v internal/service/s_category_test.go -json > ./target/log/s_category_test$(YMD).log; \
	go test -v internal/model/mod_task_test.go -json > ./target/log/mod_task_test$(YMD).log; \
	go test -v internal/service/s_task_test.go -json > ./target/log/s_task_test$(YMD).log; \
	make migrate-test-down; \
	make clean-logs; \
	end_time=$$(date +%s); \
	total_duration=$$((end_time - start_time)); \
	echo "Total execution time: $${total_duration}s";

go-test-single:
	go test -v internal/service/s_task_test.go;

go-test-ci:
	@set -e; \
	start_time=$$(date +%s); \
	migrate -path ./internal/migrations -database "$(DB)://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/test_db?sslmode=disable" up; \
	go test -v -short ./...; \
	end_time=$$(date +%s); \
	total_duration=$$((end_time - start_time)); \
	echo "Total execution time: $${total_duration}s";

.PHONY: proto grpcui migrate-create migrate-up migrate-down migrate-test-up migrate-test-down clean-logs go-test go-test-single go-test-ci
