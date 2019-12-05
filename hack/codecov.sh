#!/bin/bash
#
# Export test coverage profile on codecov.io website.
#

set -e

# codecov.io project token
CODECOV_TOKEN=${CODECOV_TOKEN:-}

# codecov script location, using app's build directory
OUTPUT_DIR=${OUTPUT_DIR:-}
CODECOV_BIN="${OUTPUT_DIR}/codecov.sh"

# directory with coverage profile files
COVERAGE_DIR=${COVERAGE_DIR:-}

# extra information about pull-request
PR_COMMIT=${PR_COMMIT:-}
PR_NUMBER=${PR_NUMBER:-}

function die() {
    echo "[ERROR] ${@}" 1>&2
    exit 1
}

function download_codecov() {
    curl --silent --output ${CODECOV_BIN} https://codecov.io/bash
    chmod +x ${CODECOV_BIN}
}

#
# Preparation
#

[[ -z "${CODECOV_TOKEN}" ]] && die "Can't codecov token!'"

[[ ! -d "${OUTPUT_DIR}" ]] && die "Can't find output directory '${OUTPUT_DIR}'"
[[ ! -f "${CODECOV_BIN}" ]] && download_codecov

[[ ! -d "${COVERAGE_DIR}" ]] && die "Can't find coverage directory '${COVERAGE_DIR}'"

#
# Composing Command
#

ARGS=(-Z)

[[ ! -z "${PR_COMMIT}" ]] && ARGS+=(-C "${PR_COMMIT}")
[[ ! -z "${PR_NUMBER}" ]] && ARGS+=(-P "${PR_NUMBER}")

export CODECOV_TOKEN

set -x
exec ${CODECOV_BIN} ${ARGS[@]} 2>&1 |tee ${OUTPUT_DIR}/codecov.log
