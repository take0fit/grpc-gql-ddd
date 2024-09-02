OS   := $(shell uname | awk '{print tolower($$0)}')
ARCH := $(shell case $$(uname -m) in (x86_64) echo amd64 ;; (aarch64) echo arm64 ;; (*) echo $$(uname -m) ;; esac)

BUF_VERSION                := 1.32.2
PROTOC_GEN_GO_VERSION      := 1.34.2
PROTOC_GEN_GO_GRPC_VERSION := 1.4.0

BIN_DIR := $(shell pwd)/bin

GQLGEN             := $(abspath $(BIN_DIR)/gqlgen)
BUF                := $(abspath $(BIN_DIR)/buf)
PROTOC_GEN_GO      := $(abspath $(BIN_DIR)/protoc-gen-go)
PROTOC_GEN_GO_GRPC := $(abspath $(BIN_DIR)/protoc-gen-go-grpc)

CACHE_UPDATE_INTERVAL ?= 30
DOCKER_COMPOSE_FILE = docker/docker-compose.yaml

gqlgen: $(GQLGEN)
$(GQLGEN):
	@cd ./tools && CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -extldflags -static" -o $(GQLGEN) github.com/99designs/gqlgen

buf: $(BUF)
$(BUF):
	@curl -sSL "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-$(shell uname -s)-$(shell uname -m)" -o $(BUF) && chmod +x $(BUF)

protoc-gen-go: $(PROTOC_GEN_GO)
$(PROTOC_GEN_GO):
	@curl -sSL https://github.com/protocolbuffers/protobuf-go/releases/download/v$(PROTOC_GEN_GO_VERSION)/protoc-gen-go.v$(PROTOC_GEN_GO_VERSION).$(OS).$(ARCH).tar.gz | tar -C $(BIN_DIR) -xzv protoc-gen-go

protoc-gen-go-grpc: $(PROTOC_GEN_GO_GRPC)
$(PROTOC_GEN_GO_GRPC):
	@curl -sSL https://github.com/grpc/grpc-go/releases/download/cmd%2Fprotoc-gen-go-grpc%2Fv$(PROTOC_GEN_GO_GRPC_VERSION)/protoc-gen-go-grpc.v$(PROTOC_GEN_GO_GRPC_VERSION).$(OS).$(ARCH).tar.gz | tar -C $(BIN_DIR) -xzv ./protoc-gen-go-grpc

.PHONY: gen-graphql
gen-graphql:
	@cd ./api && $(GQLGEN) --config ../gqlgen.yaml

.PHONY: gen-proto
gen-proto: $(BUF) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC)
	@$(BUF) generate --path ./api/geo

.PHONY: build-local
build-local:
	@echo "ローカル環境でのビルドを行います..."
	@go build -o ./bin/gql_server ./cmd/gql_server/main.go
	@go build -o ./bin/grpc_server ./cmd/geo_server/main.go

.PHONY: go-mod-tidy
go-mod-tidy:
	@echo "依存関係を整理しています..."
	go mod tidy

.PHONY: go-mod-tidy run-local
run-local: build-local
	@echo "ローカル環境でのサーバー起動を行います..."
	@export GRPC_SERVER_HOST=localhost; \
	export CACHE_UPDATE_INTERVAL=$(CACHE_UPDATE_INTERVAL); \
	./bin/gql_server &

	@export GRPC_SERVER_HOST=localhost; \
	export CACHE_UPDATE_INTERVAL=$(CACHE_UPDATE_INTERVAL); \
	./bin/grpc_server

	@echo "GraphQLサーバーとgRPCサーバーがローカルで起動しました。"

.PHONY: stop-local
stop-local:
	@echo "ローカルで起動したサーバーを停止します..."
	@pkill gql_server
	@pkill grpc_server
	@echo "サーバーを停止しました。"

.PHONY: build
build: copy-env
	docker-compose -f $(DOCKER_COMPOSE_FILE) build gql-server grpc-server

.PHONY: build-debug
build-debug: copy-env
	docker-compose -f $(DOCKER_COMPOSE_FILE) build gql-server-debug grpc-server

.PHONY: set-interval
set-interval:
	@echo "キャッシュ更新間隔を $(CACHE_UPDATE_INTERVAL) 秒に設定します"

.PHONY: run-docker
run-docker: copy-env set-interval
	CACHE_UPDATE_INTERVAL=$(CACHE_UPDATE_INTERVAL) docker-compose -f $(DOCKER_COMPOSE_FILE) up gql-server grpc-server

.PHONY: run-debug
run-debug: copy-env set-interval
	CACHE_UPDATE_INTERVAL=$(CACHE_UPDATE_INTERVAL) docker-compose -f $(DOCKER_COMPOSE_FILE) up gql-server-debug grpc-server

.PHONY: down
down:
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

.PHONY: copy-env
copy-env:
	@if [ -f .env.example ]; then \
		cp .env.example .env; \
		echo ".env.example から .env にコピーしました"; \
	else \
		echo ".env.example が見つかりません。"; \
		exit 1; \
	fi

.PHONY: moq
moq:
	@echo "Generating mocks using moq..."
	go install github.com/matryer/moq@v0.3.4
	go generate ./internal/domain/...
	@echo "Mocks generated successfully."

.PHONY: test
test: moq
	@echo "Running tests..."
	go test ./... -v
