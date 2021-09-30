# build client
cd ../cmd/grpc/client
go build

# go back to script directory
cd - > /dev/null

# move to root
cd ../

#cmd/grpc/client/client -id=resch@ocg.at -scheme=$1
cmd/grpc/client/client -id=".ch" -target="email" -from-end="3" -scheme=$1
