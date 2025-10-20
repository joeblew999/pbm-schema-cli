SHELL := /bin/bash
BIN_DIR := bin
PBHOST := $(BIN_DIR)/pbhost
PBM    := $(BIN_DIR)/pbm

PB_BIND ?= 127.0.0.1:8090
PB_DIR  ?= .pc/pb
PB_ADMIN_EMAIL ?= admin@example.com
PB_ADMIN_PASSWORD ?= admin123

BUNDLE_DIR ?= testdata/bundles/v0.1.0

.PHONY: all deps build clean fmt up run-pb token plan apply serve unit e2e test

all: build

deps:
	@echo "==> fetching deps"
	go mod tidy

build: deps
	@echo "==> building pbhost and pbm"
	mkdir -p $(BIN_DIR)
	go build -o $(PBHOST) ./cmd/pbhost
	go build -o $(PBM)    ./cmd/pbm

clean:
	rm -rf $(BIN_DIR) .pc
	go clean ./...

fmt:
	go fmt ./...

run-pb: build
	@echo "==> starting embedded PocketBase on $(PB_BIND)"
	PB_ADMIN_EMAIL=$(PB_ADMIN_EMAIL) PB_ADMIN_PASSWORD=$(PB_ADMIN_PASSWORD) \n	$(PBHOST) --http $(PB_BIND) --dir $(PB_DIR)

# expects jq installed
token:
	@./scripts/admin-token.sh $(PB_BIND) $(PB_ADMIN_EMAIL) $(PB_ADMIN_PASSWORD)

plan: build
	@echo "==> planning against PB @ $(PB_BIND) using fixtures $(BUNDLE_DIR)"
	TOKEN=$$(./scripts/admin-token.sh $(PB_BIND) $(PB_ADMIN_EMAIL) $(PB_ADMIN_PASSWORD)); \n	$(PBM) -cmd plan \n	  --url http://$(PB_BIND) --token $$TOKEN \n	  --fixtures-bundle-dir $(BUNDLE_DIR)

apply: build
	@echo "==> applying against PB @ $(PB_BIND) using fixtures $(BUNDLE_DIR)"
	TOKEN=$$(./scripts/admin-token.sh $(PB_BIND) $(PB_ADMIN_EMAIL) $(PB_ADMIN_PASSWORD)); \n	$(PBM) -cmd apply \n	  --url http://$(PB_BIND) --token $$TOKEN \n	  --fixtures-bundle-dir $(BUNDLE_DIR)

serve: build
	$(PBM) -cmd serve -bind :8088

unit: build
	go test ./pkg/... -v

# e2e with real pbhost
e2e: build
	@echo "==> e2e (starts pbhost, then plan/apply)"
	# Start PB in background
	PB_ADMIN_EMAIL=$(PB_ADMIN_EMAIL) PB_ADMIN_PASSWORD=$(PB_ADMIN_PASSWORD) \n	$(PBHOST) --http $(PB_BIND) --dir $(PB_DIR) >.pc/pb.log 2>&1 & echo $$! > .pc/pb.pid
	@./scripts/wait-healthy.sh $(PB_BIND)
	# Run plan/apply
	$(MAKE) plan
	$(MAKE) apply
	# Stop PB
	@if [ -f .pc/pb.pid ]; then kill $$(cat .pc/pb.pid) || true; fi

up: build
	process-compose up
