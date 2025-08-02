.PHONY: run
run: db
	./db

db: *.go go.*
	go build -o db .
