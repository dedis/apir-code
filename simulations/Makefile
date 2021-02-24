.PHONY: install lint 

install:
	go get -u -t ./...

lint:
	golint ./...

run_simul: 
	env GOGC=10000 go run . -config=$(config)
