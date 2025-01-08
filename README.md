# prism-test

A testing project for PRISM, a Provable Data Possession (PDP) system, designed for functionality verification and conducting experiments.

## Features

- Comprehensive testing environment for PRISM
- Supports functionality verification and reproducible experiments
- Integrates with **prism-go** and **prism-sol**

## Technologies Used

- [**prism-go**](https://github.com/prism-pdp/prism-go): a cryptographic library for PRISM
- [**prism-sol**](https://github.com/prism-pdp/prism-sol): a Solidity project for PRISM
- [**Docker**](https://www.docker.com/): For environment virtualization and management

## Installation

### 1. Install Docker

Follow the [official guide](https://docs.docker.com/get-docker/) to install Docker.

### 2. Clone the repository

```bash
git clone https://github.com/prism-pdp/prism-test.git
```

### 3. Build the Docker image

```bash
make build-img
```

## Testing

You can run simulation test with following command.
It does not run a blockchain network and manages data in json format files.
The smart contract functions are simulated by a Go program.

```bash
make simcheck
```

You can run test with following command using a local testnet of Ethereum.
It does run a local testnet and manages data by Ethereum smart contracts.

```bash
make ethcheck
```

## Running Experiments

You can run the experiments with following command.

```bash
make eval
```

