#!/usr/bin/env bash
#
# Installs libgit2 from source-code, after downloading and preparing the system to compile it.
#

set -eu

LIBGIT_HOST="github.com"
LIBGIT_PATH="libgit2/libgit2/archive/refs/tags"

LIBGIT_VERSION=${LIBGIT_VERSION:-1.2.0}

LIBGIT_TARBALL="v${LIBGIT_VERSION}.tar.gz"
LIBGIT_DIR="libgit2-${LIBGIT_VERSION}"
LIBGIT_PREFIX="${LIBGIT_PREFIX:-/usr/local}"

TMPDIR="$(mktemp -d)"
cd ${TMPDIR}

echo "# Installing dependencies..."
apt install --yes --quiet \
    curl \
    cmake \
    libssl-dev \
    python

echo "# Downloading libgit2 source-code..."
curl --silent --location --remote-name https://${LIBGIT_HOST}/${LIBGIT_PATH}/${LIBGIT_TARBALL}

echo "# Extracting tarball..."
tar xpf ${LIBGIT_TARBALL}
cd ${LIBGIT_DIR}

echo "# Building and installing libgit2!"
mkdir build && cd build
cmake .. -DCMAKE_INSTALL_PREFIX="${LIBGIT_PREFIX}"
cmake --build . --target install

echo "# Remmoving temporary director '${TMPDIR}'"
cd /tmp && rm -rf "${TMPDIR}"

echo "# Installed:"
ls -lR "${LIBGIT_PREFIX}"
