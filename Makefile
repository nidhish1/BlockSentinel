# Project variables
GO_DIR=go-listener
GO_BIN=$(GO_DIR)/blocksentinel
COMPOSE_FILE=docker-compose.yml

.PHONY: help go-build go-run go-test go-fmt go-vet go-tidy clean \
	docker-build-listener docker-build-analyzer compose-up compose-down compose-logs 

help:
	@echo "Available targets:"
	@echo "  go-build             Build Go listener binary"
	@echo "  go-run               Run Go listener (from source)"
	@echo "  go-test              Run Go tests"
	@echo "  go-fmt               Format Go code"
	@echo "  go-vet               Run go vet"
	@echo "  go-tidy              Update Go module deps"
	@echo "  docker-build-listener Build Docker image for Go listener"
	@echo "  docker-build-analyzer Build Docker image for analyzer"
	@echo "  compose-up           Start services with docker-compose"
	@echo "  compose-down         Stop services and remove containers"
	@echo "  compose-logs         Tail compose logs"
	@echo "  clean                Remove build artifacts"

# Go targets
go-build:
	cd $(GO_DIR) && go build -o blocksentinel ./...

go-run:
	cd $(GO_DIR) && go run .

go-test:
	cd $(GO_DIR) && go test ./...

go-fmt:
	cd $(GO_DIR) && go fmt ./...

go-vet:
	cd $(GO_DIR) && go vet ./...

go-tidy:
	cd $(GO_DIR) && go mod tidy

clean:
	rm -f $(GO_BIN)
	rm -f $(GO_DIR)/state.json || true

# Docker targets
docker-build-listener:
	docker build -t blocksentinel/listener:latest $(GO_DIR)

docker-build-analyzer:
	docker build -t blocksentinel/analyzer:latest analyzer

# Compose targets
dev:
	docker compose -f $(COMPOSE_FILE) up -d

compose-down:
	docker compose -f $(COMPOSE_FILE) down

compose-logs:
	docker compose -f $(COMPOSE_FILE) logs -f


