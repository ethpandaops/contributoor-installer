# Contributoor Installer

This is the installer for the Contributoor project.

## Installation

```bash
curl -O https://raw.githubusercontent.com/ethpandaops/contributoor-installer-test/refs/heads/master/install.sh && chmod +x install.sh && ./install.sh
```

## Development

### Go Tests

Execute the full test suite:

```bash
go test ./...
```

Or just run the short tests:

```bash
go test -test.short ./...
```

Or with coverage:

```bash
go test -failfast -cover -coverpkg=./... -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

### Shell Tests

You'll need [`bats`](https://github.com/bats-core/bats-core) installed if you don't already.

```bash
bats *.bats
```

If you want to run the tests with coverage, install [`kcov`](https://github.com/SimonKagstrom/kcov) and you can use the following command:

```bash
kcov --bash-parser="$(which bash)" --include-pattern=install.sh /path/to/coverage/output bats --tap install.bats
```



