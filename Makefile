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

TEST_TIMEOUT ?= 10m
TEST_FLAGS ?= -failfast -timeout=$(TEST_TIMEOUT)

UNIT_TEST_TARGET ?= ./cmd/... ./pkg/...
E2E_TEST_TARGET ?= ./test/e2e/...

COVERAGE_DIR ?= $(OUTPUT_DIR)/coverage

LIBGIT_VERSION ?= 1.3.0
LD_LIBRARY_PATH ?= /usr/local/lib

# all variables are exported to environment
.EXPORT_ALL_VARIABLES:

default: build

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
build: prepare $(OUTPUT_DIR)/$(APP)

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
	./hack/devcontainer-run.sh

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

.PHONY: lint
lint:
	@golangci-lint run

.PHONY: libgit2
libgit2:
	./hack/ubuntu/libgit2.sh