export CONFIG_PATH=configs/local.yaml
CONFIG_PATH = configs/local.yaml
TEST_CONFIG_PATH = configs/local_tests.yaml

PKGS = $(shell go list ./... | grep -v /vendor | grep -v grpc)

vet:
	@go vet $(PKGS) && echo "go vet: OK"

GOLANGCI_LINT_VERSION = v1.64.5
lint: ## Run golangci-lint
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not found. Installing $(GOLANGCI_LINT_VERSION)..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi
	@$(shell go env GOPATH)/bin/golangci-lint run ./...

fix:
	@if ! command -v fieldalignment >/dev/null 2>&1; then \
		go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest; \
	fi
	@$(shell go env GOPATH)/bin/fieldalignment -fix ./...

run:
	-@go run cmd/gateway/main.go

GEN_DIR = pkg/gen/grpc/auth
gen-proto:
	$(shell mkdir -p $(GEN_DIR))
	rm -rf $(GEN_DIR)/*.go
	@rm -rf proto/vendor
	@mkdir -p proto/vendor
	@rm -rf proto/sso/buf/*

	@buf dep update && buf generate proto
	@buf export . --output proto/vendor
	@cp -r proto/vendor/buf/ proto/sso/


# DOCKER
# Use the service name defined in docker-compose.yaml
DOCKER_APP_SERVICE = app

.PHONY: docker-up docker-down docker-reload

# Start everything
docker-up:
	docker compose up --build -d

# Stop everything
docker-down:
	docker compose down

# Rebuild the app and restart it without touching the DB
docker-reload:
	docker compose up --build -d --no-deps --force-recreate app
