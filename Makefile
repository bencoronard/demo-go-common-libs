.PHONY: test

test:
	$(LOAD_ENV) \
	go test ./... -v -cover