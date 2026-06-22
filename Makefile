.PHONY: all build proto test clean

GO := go
PROTOC := protoc
PROTOC_GEN_GO := protoc-gen-go
PROTOC_GEN_GO_GRPC := protoc-gen-go-grpc

all: proto build

proto: proto-go proto-py

proto-go:
	$(PROTOC) --go_out=. --go_opt=module=github.com/autohttp/autohttp \
		--go-grpc_out=. --go-grpc_opt=module=github.com/autohttp/autohttp \
		-I proto \
		proto/autohttp/v1/*.proto

proto-py:
	cd python/autohttp_ai && \
	$(PROTOC) --python_out=gen --pyi_out=gen \
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
	rm -rf bin/ python/autohttp_ai/gen/
