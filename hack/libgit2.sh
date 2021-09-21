#!/bin/bash
#
# Installs libgit2 package.
#

set -eu

yum install \
    --assumeyes \
    --nogpgcheck \
    --allowerasing \
    --enablerepo=rawhide \
    libgit2
