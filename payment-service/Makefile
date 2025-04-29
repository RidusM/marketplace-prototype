build:
	@go build -o ./bin/app ./cmd/main/main.go

run:
	@./bin/app

grpc:
	protoc --go_out=. --go-grpc_out=. proto/payment/payment.proto

grpc-gateway:
	protoc --go_out=. --go-grpc_out=. --grpc-gateway_out=. --grpc-gateway_opt generate_unbound_methods=true --openapiv2_out . proto/payment/payment.proto