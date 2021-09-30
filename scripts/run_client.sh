# build client
cd ../cmd/grpc/client
go build

# go back to script directory
cd - > /dev/null

# move to root
cd ../

cmd/grpc/client/client -id=resch@ocg.at -scheme=$1
