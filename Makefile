.PHONY: build install clean run

# Binary name
BINARY := cctasks

# Build directory
BUILD_DIR := bin

# Detect OS for binary extension and commands
ifeq ($(OS),Windows_NT)
	BINARY_EXT := .exe
	MKDIR := if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	RM_BUILD := if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)
	RM_ROOT := if exist $(BINARY).exe del $(BINARY).exe
else
	BINARY_EXT :=
	MKDIR := mkdir -p $(BUILD_DIR)
	RM_BUILD := rm -rf $(BUILD_DIR)
	RM_ROOT := rm -f $(BINARY)
endif

# Build the binary to bin/
build:
	@$(MKDIR)
	go build -o $(BUILD_DIR)/$(BINARY)$(BINARY_EXT) .

# Install to GOPATH/bin
install:
	go install .

# Clean build artifacts
clean:
	-@$(RM_BUILD)
	-@$(RM_ROOT)

# Run the application
run: build
	$(BUILD_DIR)/$(BINARY)$(BINARY_EXT)
