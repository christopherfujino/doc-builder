all:V: build vet

build:V:
  # https://github.com/microsoft/azurelinux/issues/3273
  go build -buildvcs=false -o db .

vet:V:
  go vet .

run:V: db
  ./db

clean:V:
  rm db
