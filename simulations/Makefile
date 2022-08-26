.PHONY: run_simul benchmarks performance

run_simul: 
	go run . -config=$(config)

single:
	$(MAKE) -s run_simul config=computationalVpir.toml; \
	$(MAKE) -s run_simul config=computationalLWE.toml; \
	$(MAKE) -s run_simul config=computationalLWE128.toml; \
	#$(MAKE) -s run_simul config=computationalPir.toml; \

preprocessing:
	$(MAKE) -s run_simul config=preprocessing.toml \
