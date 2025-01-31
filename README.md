# contributoor-installer

This repository contains the installer for the [contributoor](https://github.com/ethpandaops/contributoor) service, which collects data from Ethereum consensus clients.

## Getting Started

  ### 🔒 Installation
  Download and inspect the installation script before running:
  ```bash
  # Download the script.
  curl -O https://raw.githubusercontent.com/ethpandaops/contributoor-installer/refs/heads/master/install.sh
  
  # Inspect the script contents.
  less install.sh
  
  # Make it executable and run if you're satisfied with the contents.
  chmod +x install.sh && ./install.sh
  ```

<details>
  <summary>⚡ Quick Installation</summary>

  If you trust the source, you can run this one-liner:
  ```bash
  curl -O https://raw.githubusercontent.com/ethpandaops/contributoor-installer/refs/heads/master/install.sh && chmod +x install.sh && ./install.sh
  ```
</details>

-------------------------

  ### 🐳 With Eth-Docker

  If you're using [eth-docker](https://ethdocker.com), setup is as follows:

  - Run `./ethd update`
  - Then edit your .env file:
    - add `:contributoor.yml` to the end of `COMPOSE_FILE` variable
    - add `CONTRIBUTOOR_USERNAME` variable and set it to your username
    - add `CONTRIBUTOOR_PASSWORD` variable and set it to your password
  - Run `./ethd update`
  - Run `./ethd up`
  
  You can read more about configuring eth-docker [here](https://ethdocker.com/Usage/Advanced#specialty-yml-files).

  ### 🚀 With Rocketpool Smart Node
  
  - Install `contributoor` via the [Install Script](#-installation)
  - During the Contributoor setup:
    - Set `Beacon Node Address` to `http://eth2:5052`
    - Set `Optional Docker Network` to `rocketpool_net`
   
    Note: These can also be set later `contributoor config`
  - Run `contributoor start`

  ### ⎈ With Kubernetes (Helm)

  Contributoor can be deployed on Kubernetes using the Helm chart from the [ethereum-helm-charts](https://github.com/ethpandaops/ethereum-helm-charts) repository.

  ```bash
  # Add the Helm repository
  helm repo add ethereum-helm-charts https://ethpandaops.github.io/ethereum-helm-charts

  # Update your repositories
  helm repo update

  # Install contributoor
  helm install contributoor ethereum-helm-charts/contributoor
  ```

  For more details and configuration options, see the [contributoor chart documentation](https://github.com/ethpandaops/ethereum-helm-charts/tree/master/charts/contributoor).

### 😔 Uninstall

Uninstalling contributoor can be done by running the installer with the `-u` flag:
```bash
curl -O https://raw.githubusercontent.com/ethpandaops/contributoor-installer/refs/heads/master/install.sh && chmod +x install.sh && ./install.sh -u
```

## ⚙️ Post-Installation

> **Note:** you may need to start a new shell session before you can run the `contributoor` command.

After installation, Contributoor can be managed using these commands:

```bash
contributoor start    # Start the service
contributoor stop     # Stop the service
contributoor status   # Check service status
contributoor restart  # Restart the service
contributoor config   # View/edit configuration
contributoor update   # Update to latest version
contributoor logs     # Show logs
```

If you chose to install contributoor under a custom directory, you will need to specify the directory when running the commands, for example:

```bash
contributoor --config-path /path/to/contributoor start
```

## 🔨 Development

<details>
  <summary>Go Tests</summary>

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
</details>

<details>
  <summary>Shell Tests</summary>

  Requires [`bats`](https://github.com/bats-core/bats-core):

  ```bash
  bats *.bats
  ```

  For test coverage (requires [`kcov`](https://github.com/SimonKagstrom/kcov)):

  ```bash
  kcov --bash-parser="$(which bash)" --include-pattern=install.sh /path/to/coverage/output bats --tap install.bats
  ```
</details>

## 🤝 Contributing

Contributoor is part of EthPandaOps' suite of tools for Ethereum network operations. Contributions are welcome! Please check our [GitHub repository](https://github.com/ethpandaops) for more information.
