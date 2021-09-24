APP ?= chart-streams
MODULE = $(subst -,,$(APP))

IMAGE = quay.io/otaviof/$(APP)
IMAGE_TAG = $(IMAGE):latest
IMAGE_DEV_TAG = $(IMAGE)-dev:latest

OUTPUT_DIR ?= _output

DEVCONTAINER_ARGS ?=
RUN_ARGS ?= serve
TEST_ARGS ?=

WORKING_DIR ?= /var/tmp/chart-streams
COMMON_FLAGS ?= -v -mod=vendor

CHARTS_REPO_ARCHIVE ?= test/charts-repo.tar.gz
CHARTS_REPO_DIR ?= $(OUTPUT_DIR)/charts-repo

TEST_TIMEOUT ?= 3m
TEST_FLAGS ?= -failfast -timeout=$(TEST_TIMEOUT)

UNIT_TEST_TARGET ?= ./cmd/... ./pkg/...
E2E_TEST_TARGET ?= ./test/e2e/...

CODECOV_TOKEN ?=
COVERAGE_DIR ?= $(OUTPUT_DIR)/coverage

# all variables are exported to environment
.EXPORT_ALL_VARIABLES:

default: build

# initialize Go modules vendor directory
.PHONY: vendor
vendor:
	@go mod vendor

# clean up build directory
.PHONY: clean
clean:
	@rm -rf $(OUTPUT_DIR)

# compress local charts test repository
archive-charts-repo:
	tar zcvpf $(CHARTS_REPO_ARCHIVE) $(CHARTS_REPO_DIR)

# uncompress test charts repository archive tarball
unarchive-charts-repo:
	@rm -rf "$(CHARTS_REPO_DIR)"
	@tar zxpf $(CHARTS_REPO_ARCHIVE)

# create build and coverage directories
.PHONY: prepare
prepare: unarchive-charts-repo
	@mkdir -p $(COVERAGE_DIR) > /dev/null 2>&1 || true

# build application command-line
build: prepare vendor $(OUTPUT_DIR)/$(APP)

# application binary
$(OUTPUT_DIR)/$(APP):
	go build $(COMMON_FLAGS) -o="$(OUTPUT_DIR)/$(APP)" cmd/$(MODULE)/*

# installs all development dependencies in the development container
devcontainer-deps:
	./hack/fedora.sh
	./hack/golang.sh
	./hack/libgit2.sh
	./hack/libgit2-devel.sh
	./hack/yum-clean-up.sh
	./hack/helm.sh
# build devcontainer image
devcontainer-image:
	docker build --tag="$(IMAGE_DEV_TAG)" --file="Dockerfile.dev" .

# execute devcontainer mounting local project directory
devcontainer-run:
	docker run \
		--rm \
		--interactive \
		--tty \
		--env TMPDIR=/src \
		--volume="${PWD}:/src/$(APP)" \
		--workdir="/src/$(APP)" \
		$(IMAGE_DEV_TAG) $(DEVCONTAINER_ARGS)

# start a bash shell in devcontainer
devcontainer-exec:
	@docker exec --interactive --tty --workdir="/workspaces/$(APP)" $(APP) bash

# installs final application image dependencies
image-deps:
	./hack/fedora.sh
	./hack/libgit2.sh
	./hack/yum-clean-up.sh

# build container image with Docker
image:
	docker build --tag="$(IMAGE_TAG)" .

# execute "go run" against cmd
run:
	@test -d $(WORKING_DIR) && rm -rf $(WORKING_DIR)
	@mkdir $(WORKING_DIR)
	go run $(COMMON_FLAGS) cmd/$(MODULE)/* $(RUN_ARGS) --working-dir=$(WORKING_DIR)

# running all test targets
test: test-unit test-e2e

# run unit tests
.PHONY: test-unit
test-unit: prepare
	go test \
		$(COMMON_FLAGS) \
		$(TEST_FLAGS) \
		-coverprofile=$(COVERAGE_DIR)/coverage-unit.txt \
		$(TEST_ARGS) \
		$(UNIT_TEST_TARGET) \

# run end-to-end tests
.PHONY: test-e2e
test-e2e: prepare
	go test \
		$(COMMON_FLAGS) \
		$(TEST_FLAGS) \
		-coverprofile=$(COVERAGE_DIR)/coverage-e2e.txt \
		$(E2E_TEST_TARGET)

# codecov.io test coverage report
codecov:
	./hack/codecov.sh

.PHONY: lint
lint:
	@golangci-lint run
