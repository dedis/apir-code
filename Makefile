.PHONY: install lint keys test

PROTO_PB=lib/proto/vpir.pb.go

install:
	go get -u -t ./...

lint:
	golint ./...

run_server: $(PROTO_PB)
	cd cmd/grpc/server && go build
	go run cmd/grpc/server/main.go -id=$(id) -files=$(files) -scheme=$(scheme)

run_client: $(PROTO_PB)
	cd cmd/grpc/client && go build main.go demo.go
	cmd/grpc/client/main -id=$(id) -scheme=$(scheme)

test: $(PROTO_PB)
	go test

keys:
	cd data && go build -o parser
	cd data && ./parser

$(PROTO_PB): lib/proto/vpir.proto
	protoc --go_out=. --go_opt=paths=source_relative \
    	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	lib/proto/vpir.proto

