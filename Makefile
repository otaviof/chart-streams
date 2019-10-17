default: build

APP ?= chart-streams
GO_COMMON_FLAGS ?= -v -mod=vendor
OUTPUT_DIR ?= output

build:
	@mkdir $(OUTPUT_DIR) > /dev/null 2>&1 || true
	go build $(GO_COMMON_FLAGS) -o $(OUTPUT_DIR)/$(APP) cmd/$(APP)/*

test-unit:
	go test $(GO_COMMON_FLAGS) -timeout=1m ./...
