# build server
cd ../cmd/grpc/server
go build

# go back to simultion directory
cd - > /dev/null

# move to root
cd ../

cmd/grpc/server/server -id=$1 -files=$2 -scheme=$3
