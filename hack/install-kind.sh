#!/bin/bash
#
# Deploy kubectl and KinD, adding configuration to use KinD as the default cluster.
#

set -eu

KUBECTL_VERSION=${KUBECTL_VERSION:-}
KUBECTL_URL="https://storage.googleapis.com"
KUBECTL_URL_PATH="kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl"

KUBECTL_TARGET_DIR="${KUBECTL_TARGET_DIR:-/home/travis/bin}"
KUBECTL_BIN="${KUBECTL_TARGET_DIR}/kubectl"

function die () {
    echo "[ERROR] ${*}" 1>&2
    exit 1
}

#
# kubectl
#

[[ -z "${KUBECTL_VERSION}" ]] && die "Can't find KUBECTL_VERSION'!"

curl --location --output "${KUBECTL_BIN}" "${KUBECTL_URL}/${KUBECTL_URL_PATH}"
chmod +x "${KUBECTL_BIN}"

#
# KinD
#

go get sigs.k8s.io/kind
kind create cluster

#
# Cluster Config
#

$KUBECTL_BIN config use-context kind-kind
