gen:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto

test:
	go test -cover -race ./...

server:
	go run cmd/server/main.go -port 8787

client:
	go run cmd/client/main.go -address 0.0.0.0:8787