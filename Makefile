.PHONY: run
run: doc-builder
	./doc-builder

doc-builder:
	go build .
