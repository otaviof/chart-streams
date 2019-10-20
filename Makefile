APP ?= chart-streams
GO_COMMON_FLAGS ?= -v -mod=vendor
OUTPUT_DIR ?= output
RUN_ARGS ?= "serve"
TEST_TIMEOUT ?= 1m

default: build

# initialize Go modules vendor directory
vendor:
	go mod vendor

# build application command-line
build: vendor
	@mkdir $(OUTPUT_DIR) > /dev/null 2>&1 || true
	go build $(GO_COMMON_FLAGS) -o $(OUTPUT_DIR)/$(APP) cmd/$(APP)/*

# execute "go run" against cmd
run:
	go run $(GO_COMMON_FLAGS) cmd/$(APP)/* $(RUN_ARGS)

# running all test targets
test: test-unit test-e2e

# run unit tests
test-unit:
	go test $(GO_COMMON_FLAGS) -timeout=$(TEST_TIMEOUT) ./...

# run end-to-end tests
test-e2e:
	echo "TODO: include end-to-end tests here!"