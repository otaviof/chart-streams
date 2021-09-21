#!/bin/bash
#
# Deploys latest Golang from "rawhide" repository
#

set -eu

yum install \
    --assumeyes \
    --nogpgcheck \
    --allowerasing \
    --enablerepo=rawhide \
    golang

GOLANGCI_LINT_VERSION=${GOLANGCI_LINT_VERSION:-1.42.1}
rpm -i https://github.com/golangci/golangci-lint/releases/download/v${GOLANGCI_LINT_VERSION}/golangci-lint-${GOLANGCI_LINT_VERSION}-linux-amd64.rpm