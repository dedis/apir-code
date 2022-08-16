.PHONY: run_simul benchmarks performance

run_simul: 
	env GOGC=8000 go run . -config=$(config)

point: 
	$(MAKE) -s run_simul config=pirClassic.toml; \
	$(MAKE) -s run_simul config=pirMerkle.toml \

complex:
	$(MAKE) -s run_simul config=fss.toml \

single:
	$(MAKE) -s run_simul config=computationalPir.toml; \
	$(MAKE) -s run_simul config=computationalVpir.toml; \
	$(MAKE) -s run_simul config=computationalLWE.toml; \
	$(MAKE) -s run_simul config=computationalLWE128.toml; \

multi: 
	$(MAKE) -s run_simul config=pirClassicMulti.toml; \
	$(MAKE) -s run_simul config=pirMerkleMulti.toml \

preprocessing:
	$(MAKE) -s run_simul config=preprocessing.toml \

real:
	bash real.sh
