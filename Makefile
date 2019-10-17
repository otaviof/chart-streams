APP ?= chart-streams
GO_COMMON_FLAGS ?= -v -mod=vendor
OUTPUT_DIR ?= output

default: build

# initialize Go modules vendor directory
vendor:
	go mod vendor

# build application command-line
build: vendor
	@mkdir $(OUTPUT_DIR) > /dev/null 2>&1 || true
	go build $(GO_COMMON_FLAGS) -o $(OUTPUT_DIR)/$(APP) cmd/$(APP)/*

# running all test targets
test: test-unit test-e2e

# run unit tests
test-unit:
	go test $(GO_COMMON_FLAGS) -timeout=1m ./...

# run end-to-end tests
test-e2e:
	echo "TODO"