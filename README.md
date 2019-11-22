<p align="center">
    <img alt="chart-streams logo" src="./assets/logo/chart-streams.png">
</p>
<p align="center">
    <a alt="GoReport" href="https://goreportcard.com/report/github.com/otaviof/chart-streams">
        <img alt="GoReport" src="https://goreportcard.com/badge/github.com/otaviof/chart-streams">
    </a>
<!--
    <a alt="Code Coverage" href="https://codecov.io/gh/otaviof/chart-streams">
        <img alt="Code Coverage" src="https://codecov.io/gh/otaviof/chart-streams/branch/master/graph/badge.svg">
    </a>
  -->
    <a href="https://godoc.org/github.com/otaviof/chart-streams/pkg/chart-streams">
        <img alt="GoDoc Reference" src="https://godoc.org/github.com/otaviof/chart-streams/pkg/chart-streams?status.svg">
    </a>
    <a alt="CI Status" href="https://travis-ci.com/otaviof/chart-streams">
        <img alt="CI Status" src="https://travis-ci.com/otaviof/chart-streams.svg?branch=master">
    </a>
<!--
    <a alt="Docker-Cloud Build Status" href="https://hub.docker.com/r/otaviof/chart-streams">
        <img alt="Docker-Cloud Build Status" src="https://img.shields.io/docker/cloud/build/otaviof/chart-streams.svg">
    </a>
  -->
</p>

# `chart-streams`

`chart-streams` is a thin layer on top of a Git repository to make it behave as a Helm-Charts
repository would. With the the following advantages:

- keeping all charts data in a single place
- being able to on-the-fly generate `index.yaml`;
- use `semver` to retrieve a chart in a given `commit-id`/`branch`;

The basic workflow is represented as:

<p align="center">
    <img alt="chart-streams diagram" src="./assets/diagrams/cs-diagram-1.png">
</p>