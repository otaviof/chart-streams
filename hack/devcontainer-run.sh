#!/bin/bash

TMPDIR=$(mktemp -d)

docker run \
    --rm \
    --interactive \
    --tty \
    --mount type=bind,source="${TMPDIR}",destination=/tmp \
    --volume="${PWD}:/src/${APP}" \
    --workdir="/src/${APP}" \
    ${IMAGE_DEV_TAG} ${DEVCONTAINER_ARGS}

rm -fr "${TMPDIR}"
