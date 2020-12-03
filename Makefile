PROTO_PB=lib/proto/vpir.pb.go

run_server: $(PROTO_PB)
	go run cmd/server/main.go

run_client: $(PROTO_PB)
	go run cmd/client/main.go

test: $(PROTO_PB)
	go test

$(PROTO_PB): lib/proto/vpir.proto
	protoc --go_out=plugins=grpc:. --go_opt=paths=source_relative lib/proto/vpir.proto

