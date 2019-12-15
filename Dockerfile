# On CI, use
# FROM registry.svc.ci.openshift.org/ocp/builder:golang-1.13 AS builder

FROM openshift/origin-release:golang-1.13 AS builder

ENV LANG=en_US.utf8
ENV GIT_COMMITTER_NAME devtools
ENV GIT_COMMITTER_EMAIL devtools@redhat.com
LABEL com.redhat.delivery.appregistry=true

WORKDIR /go/src/github.com/otaviof/chart-streams

# Copy only relevant things (instead of all) to speed-up the build.
COPY . .

ARG VERBOSE=2
RUN make build

#--------------------------------------------------------------------

# On CI , use
# FROM registry.svc.ci.openshift.org/ocp/ubi-minimal:7

FROM registry.access.redhat.com/ubi7/ubi-minimal

LABEL com.redhat.delivery.appregistry=true
LABEL maintainer "Devtools <devtools@redhat.com>"
LABEL author "Shoubhik Bose <shbose@redhat.com>"
ENV LANG=en_US.utf8

COPY --from=builder /go/src/github.com/otaviof/chart-streams/build/chart-streams /usr/local/bin/chart-streams

USER 10001

ENTRYPOINT [ "/usr/local/bin/chart-streams","serve" ]