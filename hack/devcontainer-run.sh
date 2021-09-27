#!/bin/bash

docker run \
    --rm \
    --interactive \
    --env GOTMPDIR=/build \
    --env CODECOV_TOKEN="${CODECOV_TOKEN}" \
    --volume="${PWD}:/src/${APP}" \
    --workdir="/src/${APP}" \
    ${IMAGE_DEV_TAG} ${DEVCONTAINER_ARGS}

