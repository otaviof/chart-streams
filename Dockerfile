FROM golang:1.14 AS builder

ENV LANG="en_US.utf8" \
    GIT_COMMITTER_NAME="devtools" \
    GIT_COMMITTER_EMAIL="devtools@redhat.com"

WORKDIR /go/src/github.com/otaviof/chart-streams

COPY . .

RUN make

#
# Application Image
#

FROM registry.access.redhat.com/ubi8/ubi-minimal

LABEL com.redhat.delivery.appregistry="true"
LABEL maintainer="Devtools <devtools@redhat.com>"
LABEL author="Devtools <devtools@redhat.com>"

ENV LANG="en_US.utf8" \
    GIN_MODE="release"

COPY --from=builder \
    /go/src/github.com/otaviof/chart-streams/build/helm-repository-service \
    /usr/local/bin/helm-repository-service

USER 10001

ENTRYPOINT [ "/usr/local/bin/helm-repository-service", "serve" ]
