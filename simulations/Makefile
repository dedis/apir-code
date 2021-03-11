.PHONY: run_simul benchmarks performance

run_simul: 
	env GOGC=8000 go run . -config=$(config) # -cpuprofile=$(config).prof

benchmarks:
	$(MAKE) -s run_simul config=vpirSingleVector.toml; \
    		$(MAKE) -s run_simul config=vpirMultiVector.toml; \
    		$(MAKE) -s run_simul config=vpirMultiVectorBlock.toml; \

performance:
	$(MAKE) -s run_simul config=vpirMultiMatrixBlock.toml; \
	$(MAKE) -s run_simul config=vpirMultiVectorBlockDPF.toml; \
	$(MAKE) -s run_simul config=pirMatrix.toml; \
	$(MAKE) -s run_simul config=pirDPF.toml; \
	$(MAKE) -s run_simul config=merkleMatrix.toml; \
	$(MAKE) -s run_simul config=merkleDPF.toml; \

merkle:
	$(MAKE) -s run_simul config=merkleMatrix.toml; \
	$(MAKE) -s run_simul config=merkleDPF.toml; \
