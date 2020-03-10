<p align="center">
    <a alt="GoReport" href="https://goreportcard.com/report/github.com/otaviof/chart-streams">
        <img alt="GoReport" src="https://goreportcard.com/badge/github.com/otaviof/chart-streams">
    </a>
    <a alt="Code Coverage" href="https://codecov.io/gh/otaviof/chart-streams">
        <img alt="Code Coverage" src="https://codecov.io/gh/otaviof/chart-streams/branch/master/graph/badge.svg">
    </a>
    <a alt="GoDoc" href="https://godoc.org/github.com/otaviof/chart-streams/pkg/chartstreams">
        <img alt="GoDoc Reference" src="https://godoc.org/github.com/otaviof/chart-streams/pkg/chartstreams?status.svg">
    </a>
    <a alt="CI Status" href="https://travis-ci.com/otaviof/chart-streams">
        <img alt="CI Status" src="https://travis-ci.com/otaviof/chart-streams.svg?branch=master">
    </a>
    <a alt="Image Build Status" href="https://quay.io/repository/otaviof/chart-streams">
        <img alt="Image Build Status" src="https://quay.io/repository/otaviof/chart-streams/status">
    </a>
</p>

# `chart-streams`

`chart-streams` is a thin layer on top of a Git repository to make it behave as a Helm-Charts
repository would. With the the following advantages:

- Promoting Git repository as source-of-authority over Helm-Charts;
- Low-friction workflow, `index.yaml` and Chart tarballs are generated dynamically;
- Allowing clients to reach branches and commit-ids, with [Semantic Versioning][semver];

The basic workflow is represented as:

<p align="center">
    <img alt="chart-streams diagram" src="./assets/diagrams/cs-diagram-1.png">
</p>

## Usage

The usage of `chart-streams` is regular Helm-Chart repository. Therefore, you can employ Helm in
command-line to interact with this repository. For instance:

```sh
helm repo add cs http://127.0.0.1:8080
helm repo update
helm search ...
```

### Container Image

The container image are stored on [quay.io/otaviof/chart-streams][quayioimage]. To run it, execute:

```sh
podman run --publish="8080:8080" quay.io/otaviof/chart-streams:latest
docker run --publish="8080:8080" quay.io/otaviof/chart-streams:latest
```

### Configuration

Configuration parameters are exposed as environment variables as well, therefore using the prefix
`CHART_STREAMS` you can combine with the option name. For instance, *clone-depth* would then
become `CHART_STREAMS_CLONE_DEPTH` as environment variables, and on command line would then become
`--clone-depth`.

The configuration options available are:

| Parameter   | Default                            | Description                                     |
|-------------|------------------------------------|-------------------------------------------------|
| repo-url    | https://github.com/helm/charts.git | git repository URL                              |
| clone-depth | 1                                  | how many commits are pulled from Git repository |
| listen-addr | 127.0.0.1:8080                     | address the application will be listening on    |

## Endpoints

In order to behave as a Helm-Charts repository, `chart-streams` exposes the following endpoints.

### `/index.yaml`

Dinamically render `index.yaml` payload, representing actual Git repository data as Helm-Charts
repository index. Helm clients are frequently downloading this payload in order to check which charts
and versions are available in the repository.

### `/chart/:name/:version`

Also generated dynamically, this endpoint exposes the "tarball" presenting the chart name (`:name`)
and version (`:version`).

## Contributing

To build this project locally you will need [GNU/Make][gnumake] and [Golang][golang] installed.
The most important targets are `make build` and `make tst`, in order to build and run project
tests (unit and integration).

```sh
make
make test
```

Additionally, consider [`.editorconfig`](./.editorconfig) file for code standards, and
[`.travis.yml`](./.travis.yml) for the continuous integration steps.

[semver]: https://helm.sh/docs/topics/chart_best_practices/conventions/
[gnumake]: https://www.gnu.org/software/make/
[golang]: https://golang.org/
[quayioimage]: https://quay.io/repository/otaviof/chart-streams
