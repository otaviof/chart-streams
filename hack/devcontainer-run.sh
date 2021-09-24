#!/bin/bash

docker run \
    --rm \
    --interactive \
    --tty \
    --env GOTMPDIR=/build \
    --volume="${PWD}:/src/${APP}" \
    --workdir="/src/${APP}" \
    ${IMAGE_DEV_TAG} ${DEVCONTAINER_ARGS}

