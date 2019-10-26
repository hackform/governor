# CMD
BIN_DIR=bin

BIN_NAME=governor
MAIN_PATH=cmd/gov/main.go
BIN_PATH=$(BIN_DIR)/$(BIN_NAME)

SETUP_BIN_NAME=govsetup
SETUP_MAIN_PATH=cmd/setup/main.go
SETUP_BIN_PATH=$(BIN_DIR)/$(SETUP_BIN_NAME)

GO=go

.PHONY: all version

all: build

version:
	@$(GO) version
	@$(GO) env

TEST_ARGS=
COVERAGE=cover.out
COVERAGE_ARGS=-covermode count -coverprofile $(COVERAGE)
BENCHMARK_ARGS=-benchtime 5s -benchmem

.PHONY: test coverage cover bench

test:
	$(GO) test $(TEST_ARGS) -cover $(COVERAGE_ARGS) ./...

coverage:
	$(GO) tool cover -html $(COVERAGE)

cover: test coverage

bench:
	$(GO) test -bench . $(BENCHMARK_ARGS)

.PHONY: fmt vet prepare

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

prepare: fmt vet

GENSRC=$(shell find . -name '*_gen.go')

.PHONY: generate gen cleangen

generate:
	$(GO) generate ./...

gen: generate fmt

cleangen:
	rm $(GENSRC)

.PHONY: dev devsetup devsetup-setup devversion clean build-bin build-setup build

dev:
	$(GO) run -ldflags "-X main.GitHash=$$(git rev-parse --verify HEAD)" $(MAIN_PATH) --config config/configdev.yaml serve

devsetup:
	$(GO) run -ldflags "-X main.GitHash=$$(git rev-parse --verify HEAD)" $(MAIN_PATH) --config config/configdev.yaml setup

devsetup-setup:
	$(GO) run -ldflags "-X main.GitHash=$$(git rev-parse --verify HEAD)" $(SETUP_MAIN_PATH) --config config/configdev.yaml setup

devversion:
	$(GO) run -ldflags "-X main.GitHash=$$(git rev-parse --verify HEAD)" $(MAIN_PATH) --config config/configdev.yaml --version

clean:
	if [ -d $(BIN_DIR) ]; then rm -r $(BIN_DIR); fi

build-bin:
	mkdir -p $(BIN_DIR)
	if [ -f $(BIN_PATH) ]; then rm $(BIN_PATH); fi
	CGO_ENABLED=0 $(GO) build -a -tags netgo -ldflags "-w -s -extldflags '-static' -X main.GitHash=$$(git rev-parse --verify HEAD)" -o $(BIN_PATH) $(MAIN_PATH)

build-setup:
	mkdir -p $(BIN_DIR)
	if [ -f $(SETUP_BIN_PATH) ]; then rm $(SETUP_BIN_PATH); fi
	CGO_ENABLED=0 $(GO) build -a -tags netgo -ldflags "-w -s -extldflags '-static' -X main.GitHash=$$(git rev-parse --verify HEAD)" -o $(SETUP_BIN_PATH) $(SETUP_MAIN_PATH)

build: clean build-bin build-setup

## docker
DOCKER_IMAGE_NAME=governor
DOCKER_VERSION=v0.2.2
DOCKER_FILE=./cmd/gov/Dockerfile
.PHONY: build-docker produp proddown devup devdown docker-clean

build-docker:
	docker build -f $(DOCKER_FILE) -t $(DOCKER_IMAGE_NAME):$(DOCKER_VERSION) -t $(DOCKER_IMAGE_NAME):latest .

produp:
	docker-compose -f dc.main.yaml -f dc.prod.yaml -f dc.compose.yaml up -d

proddown:
	docker-compose -f dc.main.yaml -f dc.prod.yaml -f dc.compose.yaml down

devup:
	docker-compose -f dc.main.yaml -f dc.dev.yaml up -d

devdown:
	docker-compose -f dc.main.yaml -f dc.dev.yaml down

docker-clean:
	if [ "$$(docker ps -q -f status=running)" ]; \
		then docker stop $$(docker ps -q -f status=running); fi
	if [ "$$(docker ps -q -f status=restarting)" ]; \
		then docker stop $$(docker ps -q -f status=restarting); fi
	if [ "$$(docker ps -q -f status=exited)" ]; \
		then docker rm $$(docker ps -q -f status=exited); fi
	if [ "$$(docker ps -q -f status=created)" ]; \
		then docker rm $$(docker ps -q -f status=created); fi

## service
SERVICE_STACK=governor
SERVICE_DEF_DIR=defs
SERVICE_DEF_NAME=$(SERVICE_DEF_DIR)/dc.$(SERVICE_STACK).yaml
.PHONY: service launch danger-land
service:
	./servicedef-gen.sh $(SERVICE_DEF_DIR) $(SERVICE_DEF_NAME)

launch:
	docker stack deploy -c $(SERVICE_DEF_NAME) $(SERVICE_STACK)

danger-land:
	docker stack rm $(SERVICE_STACK)
