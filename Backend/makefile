CMD_DIR := cmd
BIN_DIR := bin

# Find all Go files in the CMD_DIR
GO_FILES := $(wildcard $(CMD_DIR)/*)

# Define the binaries to be created
BINARIES := $(patsubst $(CMD_DIR)/%, $(BIN_DIR)/%, $(GO_FILES))

# Default target
all: $(BINARIES)

# Rule to build each binary and zip for lambda
$(BIN_DIR)/%: $(CMD_DIR)/%/main.go
	mkdir -p $(dir $@)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o $@/bootstrap $< && ~/go/bin/build-lambda-zip -o $@.zip $@/bootstrap
	
# Clean up binaries
clean:
	rm -rf $(BIN_DIR)

.PHONY: all clean