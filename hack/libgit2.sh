#!/bin/bash
#
# Installs libgit2 package.
#

set -eu

yum install \
    --assumeyes \
    --nogpgcheck \
    --allowerasing \
    libgit2
