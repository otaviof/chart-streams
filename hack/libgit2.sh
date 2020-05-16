#!/bin/bash
#
# Installs libgit2 package.
#

set -eu

LIBGIT2_VERSION="${LIBGIT2_VERSION:-0.28}"

yum install \
    --assumeyes \
    --nogpgcheck \
    --allowerasing \
    --enablerepo=rawhide \
    libgit2_${LIBGIT2_VERSION}
