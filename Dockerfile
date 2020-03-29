FROM golang:1.14 AS builder

ENV LANG="en_US.utf8" \
    GIT_COMMITTER_NAME="devtools" \
    GIT_COMMITTER_EMAIL="devtools@redhat.com"

RUN echo "deb http://http.us.debian.org/debian/ testing non-free contrib main" >> /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y -t="testing" libgit2-dev

WORKDIR /go/src/github.com/otaviof/chart-streams

COPY . .

RUN make

#
# Application Image
#

FROM fedora:latest

LABEL com.redhat.delivery.appregistry="true"
LABEL maintainer="Devtools <devtools@redhat.com>"
LABEL author="Devtools <devtools@redhat.com>"

ENV LANG="en_US.utf8" \
    GIN_MODE="release"

RUN yum install -y git libgit2-devel && \
    rm -rf /var/cache /var/log/dnf* /var/log/yum.*

COPY --from=builder \
    /go/src/github.com/otaviof/chart-streams/build/chart-streams \
    /usr/local/bin/chart-streams

USER 10001

VOLUME [ "/var/lib/chart-streams" ]

ENTRYPOINT [ "/usr/local/bin/chart-streams", "serve" ]
