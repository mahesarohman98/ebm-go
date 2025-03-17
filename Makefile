# Variables
APP_NAME := ebmgo
GOOS := linux
GOARCH := amd64
TAGS := sqlite_fts5
LD_FLAGS := -s -w -extldflags '-static'
CGO_ENABLED := 1
CC := musl-gcc

# Default target
all: clean build strip

# Build target
build:
	@echo "Building $(APP_NAME)..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) CC=$(CC) \
	go build -o $(APP_NAME) \
	-tags '$(TAGS)' \
	-ldflags="$(LD_FLAGS)" \
	-trimpath .

# Strip binary (optional)
strip:
	@echo "Stripping binary..."
	strip $(APP_NAME)

# Clean target
clean:
	@echo "Cleaning..."
	rm -f $(APP_NAME)

# Run target
run: build
	./$(APP_NAME)
