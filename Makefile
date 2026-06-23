.PHONY: all build proto proto-go proto-py test vet clean

GO := go
PATH := $(shell $(GO) env GOPATH)/bin:$(PATH)
PROTOC := protoc
PROTOC_GEN_GO := protoc-gen-go
PROTOC_GEN_GO_GRPC := protoc-gen-go-grpc
PYTHON := python3

all: proto build

proto: proto-go proto-py

proto-go:
	$(PROTOC) --go_out=. --go_opt=module=github.com/autohttp/autohttp \
		--go-grpc_out=. --go-grpc_opt=module=github.com/autohttp/autohttp \
		-I proto \
		proto/autohttp/v1/*.proto

proto-py: proto-py-ai proto-py-worker

proto-py-ai:
	cd python/autohttp_ai && \
	$(PYTHON) -m grpc_tools.protoc --python_out=gen --pyi_out=gen \
		--grpc_python_out=gen \
		-I ../../proto \
		../../proto/autohttp/v1/*.proto

proto-py-worker:
	cd python/autohttp_worker && \
	$(PYTHON) -m grpc_tools.protoc --python_out=gen --pyi_out=gen \
		--grpc_python_out=gen \
		-I ../../proto \
		../../proto/autohttp/v1/*.proto

build:
	$(GO) build -o bin/autohttp ./cmd/autohttp

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

clean:
	rm -rf bin/ python/autohttp_ai/gen/ python/autohttp_worker/gen/
