FROM golang:1.24.1-alpine AS modules

COPY ./x509 /app/x509

COPY go.sum go.mod /modules/

WORKDIR /modules

RUN go mod download

FROM golang:1.24.1-alpine AS builder

COPY --from=modules /go/pkg /go/pkg
COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o ./bin/user ./cmd/user/main.go

FROM scratch

COPY --from=builder /app/configs /configs
COPY --from=builder /app/migrations /migrations
COPY --from=builder /app/bin/user /main

CMD ["/main"]