.PHONY: build test clean run lint coverage docker-build docker-run

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
MAIN_PATH=./cmd/k8s-diff-informer
BINARY_NAME=k8s-diff-informer
DOCKER_IMAGE=k8s-diff-informer

# Default target
all: test build

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PATH)

# Run the application
run:
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PATH)
	./$(BINARY_NAME)

# Test the application
test:
	$(GOTEST) -v ./...

# Clean the build
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Install dependencies
deps:
	$(GOMOD) tidy

# Run linting
lint:
	golangci-lint run ./...

# Run tests with coverage
coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Run in Docker
docker-run:
	docker run --rm -e SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
		-e WATCHED_RESOURCE_NAMES=pods,deployments,services \
		-e WATCHED_NAMESPACES=default,kube-system \
		-v $(HOME)/.kube/config:/root/.kube/config:ro \
		$(DOCKER_IMAGE)

# Generate documentation
docs:
	godoc -http=:6060