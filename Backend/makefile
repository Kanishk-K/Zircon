CMD_DIR := cmd
BIN_DIR := bin

# Find all Go files in the CMD_DIR
GO_FILES := $(wildcard $(CMD_DIR)/*)

# Define the binaries to be created
BINARIES := $(patsubst $(CMD_DIR)/%, $(BIN_DIR)/%, $(GO_FILES))

# Define the filename for the video generation
filename ?= testNewEvent.json

# Default target
all: $(BINARIES)

# Rule to build each binary and zip for lambda
$(BIN_DIR)/%: $(CMD_DIR)/%/main.go
	mkdir -p $(dir $@)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o $@/bootstrap $< && ~/go/bin/build-lambda-zip -o $@.zip $@/bootstrap
	
# Clean up binaries
clean:
	rm -rf $(BIN_DIR)

auth:
	aws ecr-public get-login-password --region us-east-1 | docker login --username AWS --password-stdin public.ecr.aws

api:
	sam.cmd build --cached && sam.cmd local start-api --env-vars .env

invoke:
	sam.cmd build --cached && sam.cmd local start-lambda --env-vars .env

tts: 
	@[ -z $(entryID) ] && echo "entryID is required add the argument entryID=entryID" && exit 1 || echo "entryID is $(entryID)"
	sam.cmd build --cached && cat cmd/Subtitles/testEvent.json | sed 's/TEST_EVENT/$(entryID)/g' | sam.cmd local invoke --env-vars .env --event - TTSGenerationFunction

video:
	@[ -z $(entryID) ] && echo "entryID is required add the argument entryID=entryID" && exit 1 || echo "entryID is $(entryID)"
	sam.cmd build --cached && cat cmd/Queue/$(filename) | sed 's/TEST_EVENT/$(entryID)/g' | sam.cmd local invoke --docker-network zircon --env-vars .env --event - VideoGenerationFunction

.PHONY: all clean