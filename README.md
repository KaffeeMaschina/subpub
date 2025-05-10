SubPub is a gRPC service with an implementation of subscriber-publisher architecture.
It allows to subscribe to subject and to get messages dedicated to this subject. And to publish messages to subject.

This service consists of three components:
    1. EventBus. The internal logic of SubPub system. It manages subscribers, publishers and transactions between them. /mysubpub
    2. gRPC Server. It provides connection to EventBus for clients. /cmd/grpc_server
    3. gRPC Client. It is an API to interact with server. /pkg/PubSubClient

To start server:
    1. get dependencies with command 'make get-deps'
    2. configure server with local.env
    3.start local server with command 'make local-server-run'
    3. start client with command 'make local-client-run'

To regenerate Protobuf schema to Golang code:
    1. get dependencies if not yet with command 'make get-deps'
    2. install protoc-gen-go and protoc-gen-go-grpc with command 'install-deps'
    3. generate code with command 'generate'

Features

- Publish messages to specific topics
- Subscribe to topics and receive messages in real-time
- Graceful shutdown handling
- Error handling with gRPC status codes
- Environment-based configuration