FROM golang:1.22.4-alpine3.20 AS builder
ARG COMMIT_HASH=""
ARG BUILD_VERSION=""
ARG BUILD_TIME=""
ARG BRANCH=""
WORKDIR /app
COPY . .
RUN COMMIT_HASH="${COMMIT_HASH}" BUILD_VERSION="${BUILD_VERSION}" BUILD_TIME="${BUILD_TIME}" BRANCH="${BRANCH}" ./script/build.sh

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/target .
RUN apk add logrotate tzdata
COPY --from=builder /app/init/logrotate.d/go-todolist-grpc /etc/logrotate.d/go-todolist-grpc
CMD crond && ./go-todolist-grpc
