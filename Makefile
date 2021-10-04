.PHONY: install lint keys test

PROTO_PB=lib/proto/vpir.pb.go

install:
	go get -u -t ./...

lint:
	golint ./...

run_server: $(PROTO_PB)
	cd cmd/grpc/server && go build
	go run cmd/grpc/server/main.go -id=$(id) -files=$(files) -scheme=$(scheme)

build_client: $(PROTO_PB)
	cd cmd/grpc/client && go build -o client .

run_client: build_client
	cmd/grpc/client/client -id=$(id) -scheme=$(scheme)

run_demo: build_client
	cmd/grpc/client/client -demo -scheme=$(scheme)

test: $(PROTO_PB)
	go test

keys:
	cd data && go build -o parser
	cd data && ./parser

$(PROTO_PB): lib/proto/vpir.proto
	protoc --go_out=. --go_opt=paths=source_relative \
    	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	lib/proto/vpir.proto

