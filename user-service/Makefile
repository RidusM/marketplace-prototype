APP_NAME = User-Service

grpc-gen: proto/app.proto
	protoc --go_out=pkg/api/client --go_opt=paths=source_relative \
        --go-grpc_out=pkg/api/client --go-grpc_opt=paths=source_relative \
        proto/app.proto

build: cmd/user/main.go internal configs pkg proto
	go build -o ${APP_NAME} cmd/user/main.go

.PHONY: migrations-up migrations-down clean run lint

migrations-up:
	migrate -source file://migrations -database "postgres://themaxs:1234@localhost:5432/user_service?sslmode=disable" up

migrations-down:
	migrate -source file://migrations -database "postgres://themaxs:1234@localhost:5432/user_service?sslmode=disable" down

run: build
	./${APP_NAME}

clean:
	rm -f ${APP_NAME}

lint:
	golangci-lint run