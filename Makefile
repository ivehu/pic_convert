BINARY_NAME = pic_convert
PLATFORMS = linux/amd64 windows/amd64 darwin/amd64 darwin/arm64

.PHONY: all clean run build

all: clean build

build:
	@echo "Building for current platform..."
	go build -o $(BINARY_NAME)

build-multi:
	@echo "Building for multiple platforms..."
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} go build -o bin/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}; \
	done

clean:
	@echo "Cleaning binaries..."
	@rm -f $(BINARY_NAME) bin/*

run:
	@go run .