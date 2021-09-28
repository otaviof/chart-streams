#!/bin/bash

docker run \
    --rm \
    --interactive \
    --env GOTMPDIR=/build \
    --volume="${PWD}:/src/${APP}" \
    --workdir="/src/${APP}" \
    ${IMAGE_DEV_TAG} ${DEVCONTAINER_ARGS}

