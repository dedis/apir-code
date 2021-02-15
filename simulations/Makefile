.PHONY: install lint 

install:
	go get -u -t ./...

lint:
	golint ./...

run_simul: 
	go run . -config=$(config)
