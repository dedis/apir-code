# build client
cd ../cmd/grpc/client
go build

# go back to script directory
cd - > /dev/null

# move to root
cd ../

cmd/grpc/client/client -id=alex.braulio@varidi.com -scheme=$1
