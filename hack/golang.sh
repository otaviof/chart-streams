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
