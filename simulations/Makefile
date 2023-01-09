.PHONY: run_simul single preprocessing

run_simul: 
	go run . -config=$(config)

single:
	$(MAKE) -s run_simul config=computationalDH.toml; \
	$(MAKE) -s run_simul config=computationalLWE.toml; \
	$(MAKE) -s run_simul config=computationalLWE128.toml; \

preprocessing:
	$(MAKE) -s run_simul config=preprocessing.toml \
