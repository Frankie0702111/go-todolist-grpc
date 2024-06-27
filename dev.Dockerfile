FROM golang:1.22.4-alpine3.20

WORKDIR /app

RUN go install github.com/cosmtrek/air@v1.40.4; \
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1;

CMD /go/bin/air
