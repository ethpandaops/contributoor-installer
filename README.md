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

### Shell Tests

You'll need [`bats`](https://github.com/bats-core/bats-core) installed if you don't already.

```bash
bats *.bats
```
