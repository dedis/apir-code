.PHONY: install lint run_simul benchmarks performance

install:
	go get -u -t ./...

lint:
	golint ./...

run_simul: 
	env GOGC=10000 go run . -config=$(config)

benchmarks:
	$(MAKE) -s run_simul config=vpirSingleVector.toml; \
    $(MAKE) -s run_simul config=vpirMultiVector.toml; \
    $(MAKE) -s run_simul config=vpirMultiVectorBlock.toml; \

performance:
	$(MAKE) -s run_simul config=vpirMultiMatrixBlock.toml; \
    $(MAKE) -s run_simul config=vpirMultiVectorBlockDPF.toml; \
    $(MAKE) -s run_simul config=pirMatrix.toml; \
    $(MAKE) -s run_simul config=pirDPF.toml; \
