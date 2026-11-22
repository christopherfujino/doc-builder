all:V: build vet

build:V:
  go build -o db .

vet:V:
  go vet .

run:V: db
  ./db

clean:V:
  rm db
