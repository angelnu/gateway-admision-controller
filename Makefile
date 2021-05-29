
# Image URL to use all building/pushing image targets
IMG ?= template-container-image:latest

# Run tests
test: build
	go test -v .

# Build manager binary
build: fmt vet
	go build -o app main.go

# Download dependencies
download:
	go mod download

# Download dependencies
tidy: download
	go mod tidy

# Run go fmt against code
fmt: tidy
	go fmt ./...

# Run go vet against code
vet: tidy
	go vet ./...

# Build the docker image
docker-build:
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}