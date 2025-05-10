LOCAL_BIN:=$(CURDIR)/bin

install-deps:
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.4
	GOBIN=$(LOCAL_BIN) go install -mod=mod google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

get-deps:
	go get -u google.golang.org/protobuf/cmd/protoc-gen-go
	go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc

	go get github.com/joho/godotenv

local-server-run:
	go run cmd/grpc_server/main.go -config-path=local.env

local-client-run:
	go run cmd/grpc_client/main.go -config-path=local.env

generate:
	make generate-pubsub-api

generate-pubsub-api:
	mkdir -p pkg/bus_v1
	protoc --proto_path api/bus_v1 \
	--go_out=pkg/bus_v1 --go_opt=paths=source_relative \
	--plugin=protoc-gen-go=bin/protoc-gen-go \
	--go-grpc_out=pkg/bus_v1 --go-grpc_opt=paths=source_relative \
	--plugin=protoc-gen-go-grpc=bin/protoc-gen-go-grpc \
	api/bus_v1/bus.proto