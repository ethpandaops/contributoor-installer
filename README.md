# Contributoor - Installation

This repository contains the installer for the [Contributoor](https://github.com/ethpandaops/contributoor) service, which collects data from Ethereum consensus clients.

## Getting Started

```bash
curl -O https://raw.githubusercontent.com/ethpandaops/contributoor-installer/refs/heads/master/install.sh && chmod +x install.sh && ./install.sh
```
> If you would prefer to check the installation script before running it, you may download and run it manually.

## Post-Installation

> **Note:** you may need to start a new shell session before you can run the `contributoor` command.

After installation, Contributoor can be managed using these commands:

```bash
contributoor start    # Start the service
contributoor stop     # Stop the service
contributoor status   # Check service status
contributoor restart  # Restart the service
contributoor config   # View/edit configuration
contributoor update   # Update to latest version
```

## Development

### Go Tests

Execute the full test suite:

```bash
go test ./...
```

Run short tests only:

```bash
go test -test.short ./...
```

Run with coverage:

```bash
go test -failfast -cover -coverpkg=./... -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

### Shell Tests

Requires [`bats`](https://github.com/bats-core/bats-core):

```bash
bats *.bats
```

For test coverage (requires [`kcov`](https://github.com/SimonKagstrom/kcov)):

```bash
kcov --bash-parser="$(which bash)" --include-pattern=install.sh /path/to/coverage/output bats --tap install.bats
```

## Contributing

Contributoor is part of EthPandaOps' suite of tools for Ethereum network operations. Contributions are welcome! Please check our [GitHub repository](https://github.com/ethpandaops) for more information.
