.PHONY: run_simul benchmarks performance

run_simul: 
	go run . -config=$(config)

single:
	$(MAKE) -s run_simul config=computationalPir.toml; \
	$(MAKE) -s run_simul config=computationalVpir.toml; \
	$(MAKE) -s run_simul config=computationalLWE.toml; \

preprocessing:
	$(MAKE) -s run_simul config=preprocessing.toml \
