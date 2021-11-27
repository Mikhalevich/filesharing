all: build

build:
	go build -mod=vendor -o ./bin/filesharing cmd/filesharing/main.go

up:
	./scripts/run.sh

vendor:
	go mod tidy
	go mod vendor

proto:
	./scripts/update_proto.sh
