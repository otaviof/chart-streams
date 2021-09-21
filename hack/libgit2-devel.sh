#!/bin/bash
#
# Installs libgit2 development package.
#

set -eu

yum install \
    --assumeyes \
    --nogpgcheck \
    --allowerasing \
    --enablerepo=rawhide \
    libgit2-devel
