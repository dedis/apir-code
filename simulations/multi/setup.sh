export GOGC=8000

# build server
cd server
go build 
cd ..

# build client
cd client
go build 
cd ..
