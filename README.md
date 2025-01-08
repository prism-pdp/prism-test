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

## System Requirements

This project has been tested and is verified to work on **Ubuntu 24.04.01 LTS**.
We recommend using this version of Ubuntu for the best experience.
Other operating systems or versions might require additional setup or may not be supported.

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

You can run all experiments with following command.

```bash
make experiment
```

This experiments consist of the following four experiments.

- experiment-gentags: Experiment of average processing time for generating tags
- experiment-auditing: Experiment of average processing time for auditing
- experiment-contract: Experiment of average gas consumption of the smart contract
- experiment-frequency: Experiment of detection capability for data corruption

If you want to run individual experiment, you can run it with following commands.

```bash
make experiment-gentags
```

```bash
make experiment-auditing
```

```bash
make experiment-contract
```

```bash
make experiment-frequency
```

