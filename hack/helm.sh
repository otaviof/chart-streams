#!/bin/bash

set -ex

HELM_VERSION=${HELM_VERSION:-3.7.0}
HELM_OS=${GOOS:-linux}
HELM_ARCH=${GOARCH:-amd64}

KEEP_TMPDIR=${KEEP_TMPDIR:-0}
TMPDIR="$(mktemp -d)"

curl -s --output-dir "${TMPDIR}" -O https://get.helm.sh/helm-v${HELM_VERSION}-${HELM_OS}-${HELM_ARCH}.tar.gz
tar -C "${TMPDIR}" -xf "${TMPDIR}/helm-v${HELM_VERSION}-${HELM_OS}-${HELM_ARCH}.tar.gz" 
install -v "${TMPDIR}/${HELM_OS}-${HELM_ARCH}/helm" /usr/local/bin/

if [ $KEEP_TMPDIR -eq 0 ]; then
    rm -fr "${TMPDIR}"
fi