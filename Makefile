.PHONY: build install clean

# Binary name
BINARY := cctasks

# Build directory
BUILD_DIR := bin

# Detect OS for binary extension
ifeq ($(OS),Windows_NT)
	BINARY_EXT := .exe
else
	BINARY_EXT :=
endif

# Build the binary to bin/
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY)$(BINARY_EXT) .

# Install to GOPATH/bin
install:
	go install .

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY) $(BINARY)$(BINARY_EXT)

# Run the application
run: build
	./$(BUILD_DIR)/$(BINARY)$(BINARY_EXT)
