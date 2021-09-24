#!/bin/bash
#
# Installs dependencies needed in Fedora, before running other scripts.
#

set -eu

yum install --assumeyes \
    fedora-repos-rawhide

yum makecache
