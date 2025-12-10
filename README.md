# mddiff

Media Directory Diffing CLI

## Usage

```sh
mddiff path/to/dir1 path/to/dir2
```

## Contributing


### Prerequisites

- golangci-lint - run multiple linters in parallel

```sh
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.7.1
```

- (Optional) Cobra CLI - for adding new CLI commands

```sh
go install github.com/spf13/cobra-cli@latest
```


### Development Process

- Check linting and formatting

```sh
make lint
```

- Autofix linting and formatting

```sh
make fmt
```

- Run tests

```sh
make test
```

- Build the app for all supported platforms

```sh
# Generate a local dev build
make build
# Generate a release build - this is handled by GitHub Actions
MDDIFF_VERSION=x.x.x make build
```

- Clean up

```sh
make clean
```


### Developing CI Actions

- Install the Act CLI to test GitHub Actions locally

```sh
brew install act
```

- Test the release GitHub Actions workflow

```sh
act release -e test/event.json
```
