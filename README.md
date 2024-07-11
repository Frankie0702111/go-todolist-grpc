# Project description
This is a To-Do List project implemented using Go language and gRPC.

# Table of Contents
 - [Software requirements](#software-requirements)
 - [Project plugins](#project-plugins)
 - [How to build project](#how-to-build-project)
 - [Folder structure](#folder-structure)
 - [Folder definition](#folder-definition)

# Software requirements
 - **Compilation tools**
    - [Vscode](https://code.visualstudio.com/)
 - **Database**
    - [PostgreSQL](https://aws.amazon.com/tw/rds/): v15.7
 - **Programming language**
    - [Go](https://go.dev/dl/): v1.22
 - **Deveops**
    - [Docker GUI](https://www.docker.com/products/docker-desktop/)
 - **Other**
    - [Protoc](https://grpc.io/docs/protoc-installation/)
    - [Postman](https://www.postman.com/downloads/)

# Project plugins
- **Server**
    - [gRPC](https://github.com/grpc/grpc-go)
    - [Gateway](https://github.com/grpc-ecosystem/grpc-gateway)
- **Database**
    - [GORM](https://github.com/go-gorm/gorm)
    - [PostgreSQL](gorm.io/driver/postgres)
- **Interceptor**
    - [Middleware](https://github.com/grpc-ecosystem/grpc-gateway)
- **Other**
    - [Proto](https://github.com/protocolbuffers/protobuf-go)
    - [Go-jwt](https://github.com/golang-jwt/jwt)
    - [Viper](https://github.com/spf13/viper)
    - [Testify](https://github.com/stretchr/testify)

# How to build project
## 1.Clone GitHub project to local
```bash
git clone https://github.com/Frankie0702111/go-todolist-grpc.git
```

## 2.Set up Docker information, such as database, Redis, Server
```bash
cd go-todolist-grpc
cp app.env.example app.env

vim app.env
vim internal/config/env.go
```

## 3.Build docker image and start
```bash
# Create docker image
docker compose build --no-cache

# Run docker
docker compose up -d

# Stop docker
docker compose stop
```

## 4.Set up basic information, such as database, Redis, AWS, JWT
```bash
cp ./internal/config/env.go.example ./internal/config/env.go
vim ./internal/config/env.go
```

## 5.Generate db migrations
```bash
# Up all migration
make migrate-up

# Down all migration
make migrate-down

# Specify batch up or down (default = 3 -> user, category, task)
make migrate-up number=1
make migrate-down number=1
```

# Folder structure
```
├── Dockerfile
├── LICENSE
├── Makefile
├── README.md
├── api
│   ├── pb
│   │   ├── category.pb.go
│   │   ├── model.pb.go
│   │   ├── public.pb.go
│   │   ├── task.pb.go
│   │   ├── todolist.pb.go
│   │   ├── todolist.pb.gw.go
│   │   ├── todolist_grpc.pb.go
│   │   └── user.pb.go
│   └── proto
│       ├── category.proto
│       ├── google
│       │   └── api
│       │       ├── annotations.proto
│       │       ├── field_behavior.proto
│       │       ├── http.proto
│       │       └── httpbody.proto
│       ├── model.proto
│       ├── public.proto
│       ├── task.proto
│       ├── todolist.proto
│       └── user.proto
├── app.env.example
├── cmd
│   └── go-todolist-grpc
│       └── main.go
├── dev.Dockerfile
├── docker
│   └── pgsql
│       └── init.sql
├── docker-compose.yaml
├── go.mod
├── go.sum
├── init
│   └── logrotate.d
│       └── go-todolist-grpc
├── internal
│   ├── config
│   │   ├── config.go
│   │   ├── config_test.go
│   │   └── env.go.example
│   ├── middleware
│   │   └── authorization.go
│   ├── migrations
│   │   ├── 000001_create_users_table.down.sql
│   │   ├── 000001_create_users_table.up.sql
│   │   ├── 000002_create_categories_table.down.sql
│   │   ├── 000002_create_categories_table.up.sql
│   │   ├── 000003_create_tasks_table.down.sql
│   │   └── 000003_create_tasks_table.up.sql
│   ├── model
│   │   ├── mod_.go
│   │   ├── mod_category.go
│   │   ├── mod_category_test.go
│   │   ├── mod_task.go
│   │   ├── mod_task_test.go
│   │   ├── mod_user.go
│   │   └── mod_user_test.go
│   ├── pkg
│   │   ├── db
│   │   │   ├── builder
│   │   │   │   └── builder.go
│   │   │   ├── condition
│   │   │   │   ├── clausebuilder.go
│   │   │   │   └── condition.go
│   │   │   ├── db.go
│   │   │   ├── db_test.go
│   │   │   └── field
│   │   │       └── field.go
│   │   ├── log
│   │   │   └── log.go
│   │   └── util
│   │       ├── hash.go
│   │       ├── hash_test.go
│   │       ├── jwt.go
│   │       ├── jwt_test.go
│   │       ├── random.go
│   │       ├── random_test.go
│   │       ├── th.go
│   │       └── th_test.go
│   └── service
│       ├── s_.go
│       ├── s_category.go
│       ├── s_category_test.go
│       ├── s_take.go
│       ├── s_task_test.go
│       ├── s_user.go
│       └── s_user_test.go
├── postman
│   └── go-todolist-grpc (gateway).postman_collection.json
└── script
    └── build.sh
```

# Folder definition
- `api`: Contains Protocol Buffers definitions and generated Go code
    - `pb`: Generated Go code from Protocol Buffers
    - `proto`: Protocol Buffers definition files

- `cmd/go-todolist-grpc`: Main application for the gRPC and gateway server

- `docker`: Docker-related files
    - `pgsql`: PostgreSQL initialization scripts

- `init/logrotate.d`: Log rotation configuration

- `internal`: Internal packages and implementations
    - `config`: Configuration files and environment settings
    - `middleware`: Middleware implementations
    - `migrations`: Database migration files
    - `model`: Data model definitions
    - `pkg`: Common packages
        - `db`: Database-related operations
        - `log`: Logging utilities
        - `util`: Utility functions
    - `service`: gRPC service implementations

- `postman`: Postman collection for API testing

- `script`: Build and utility scripts
