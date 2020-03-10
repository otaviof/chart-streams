# application name
APP ?= chart-streams
# sanitizing app variable to become a valid go module name
MODULE = $(subst -,,$(APP))

RUN_ARGS ?= serve
COMMON_FLAGS ?= -v -mod=vendor

TEST_TIMEOUT ?= 3m
TEST_FLAGS ?= -failfast -timeout=$(TEST_TIMEOUT)
CODECOV_TOKEN ?=
COVERAGE_DIR ?= $(OUTPUT_DIR)/coverage

OUTPUT_DIR ?= build

KUBECTL_VERSION ?= v1.16.3

# used in `codecov.sh` script
export OUTPUT_DIR
export COVERAGE_DIR
export CODECOV_TOKEN

# used in `install-kind.sh` script
export KUBECTL_VERSION

default: build

# initialize Go modules vendor directory
.PHONY: vendor
vendor:
	go mod vendor

# clean up build directory
.PHONY: clean
clean:
	@rm -rf $(OUTPUT_DIR)

# create build and coverage directories
.PHONY: prepare
prepare:
	@mkdir -p $(COVERAGE_DIR) > /dev/null 2>&1 || true

# build application command-line
build: prepare vendor
	go build $(COMMON_FLAGS) -o="$(OUTPUT_DIR)/$(APP)" cmd/$(MODULE)/*

# execute "go run" against cmd
run:
	go run $(COMMON_FLAGS) cmd/$(MODULE)/* $(RUN_ARGS)

# invoke script to deploy kubectl and kind
kind:
	./hack/install-kind.sh

# running all test targets
test: test-unit test-e2e

# run unit tests
.PHONY: test-unit
test-unit: prepare
	go test \
		$(COMMON_FLAGS) \
		$(TEST_FLAGS) \
		-coverprofile=$(COVERAGE_DIR)/coverage-unit.txt \
		./cmd/... \
		./pkg/...

# run end-to-end tests
.PHONY: test-e2e
test-e2e:
	go test \
		$(COMMON_FLAGS) \
		$(TEST_FLAGS) \
		-coverprofile=$(COVERAGE_DIR)/coverage-e2e.txt \
		./test/...

# codecov.io test coverage report
codecov:
	./hack/codecov.sh
