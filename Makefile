# application name
APP ?= chart-streams
# sanitizing app variable to become a valid go module name
MODULE = $(subst -,,$(APP))
# container image tag
IMAGE_TAG ?= "quay.io/otaviof/$(APP):latest"
# build directory
OUTPUT_DIR ?= build

RUN_ARGS ?= serve
COMMON_FLAGS ?= -v -mod=vendor

CHARTS_REPO_ARCHIVE ?= test/charts-repo.tar.gz
CHARTS_REPO_DIR ?= $(OUTPUT_DIR)/charts-repo

TEST_TIMEOUT ?= 3m
TEST_FLAGS ?= -failfast -timeout=$(TEST_TIMEOUT)
CODECOV_TOKEN ?=
COVERAGE_DIR ?= $(OUTPUT_DIR)/coverage

# used in `codecov.sh` script
export OUTPUT_DIR
export COVERAGE_DIR
export CODECOV_TOKEN

default: build

# initialize Go modules vendor directory
.PHONY: vendor
vendor:
	go mod vendor

# clean up build directory
.PHONY: clean
clean:
	@rm -rf $(OUTPUT_DIR)

# compress local charts test repository
archive-charts-repo:
	tar zcvpf $(CHARTS_REPO_ARCHIVE) $(CHARTS_REPO_DIR)

# uncompress test charts repository archive tarball
unarchive-charts-repo:
	rm -rf "$(CHARTS_REPO_DIR)"
	tar zxvpf $(CHARTS_REPO_ARCHIVE)

# create build and coverage directories
.PHONY: prepare
prepare: unarchive-charts-repo
	@mkdir -p $(COVERAGE_DIR) > /dev/null 2>&1 || true

# build application command-line
build: prepare vendor
	go build $(COMMON_FLAGS) -o="$(OUTPUT_DIR)/$(APP)" cmd/$(MODULE)/*

# build container image with Docker
image:
	docker build --tag="$(IMAGE_TAG)" .

# execute "go run" against cmd
run:
	go run $(COMMON_FLAGS) cmd/$(MODULE)/* $(RUN_ARGS)

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
test-e2e: prepare
	go test \
		$(COMMON_FLAGS) \
		$(TEST_FLAGS) \
		-coverprofile=$(COVERAGE_DIR)/coverage-e2e.txt \
		./test/...

# codecov.io test coverage report
codecov:
	./hack/codecov.sh
