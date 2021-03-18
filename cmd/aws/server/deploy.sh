
#!/bin/bash

set -e

rm -f main.zip vpir-lambda
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on go build
zip -X -r ./main.zip vpir-lambda

aws lambda update-function-code --function-name VPIR --zip-file fileb://main.zip
