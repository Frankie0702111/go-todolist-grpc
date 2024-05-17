FROM golang:1.19.13-alpine3.18

WORKDIR /app

RUN go install github.com/cosmtrek/air@v1.40.4;

CMD /go/bin/air
