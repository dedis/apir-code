PROTO_PB=lib/proto/vpir.pb.go

run_server: $(PROTO_PB)
	cd cmd/server && go build
	go run cmd/server/main.go -id=$(id) -scheme=$(scheme)

run_client: $(PROTO_PB)
	cd cmd/client && go build
	go run cmd/client/main.go -id=$(id) -scheme=$(scheme)

test: $(PROTO_PB)
	go test

$(PROTO_PB): lib/proto/vpir.proto
	protoc --go_out=. --go_opt=paths=source_relative \
    	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	lib/proto/vpir.proto

