all: build

.PHONY: build
build:
	go build -mod=vendor -o ./bin/filesharing cmd/filesharing/main.go

.PHONY: run
run:
	./scripts/run.sh

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor

.PHONY: proto
proto:
	./scripts/update_proto.sh
