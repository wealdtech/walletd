# walletd

[![Tag](https://img.shields.io/github/tag/wealdtech/walletd.svg)](https://github.com/wealdtech/walletd/releases/)
[![License](https://img.shields.io/github/license/wealdtech/walletd.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/wealdtech/walletd?status.svg)](https://godoc.org/github.com/wealdtech/walletd)
[![Travis CI](https://img.shields.io/travis/wealdtech/walletd.svg)](https://travis-ci.org/wealdtech/walletd)
[![codecov.io](https://img.shields.io/codecov/c/github/wealdtech/walletd.svg)](https://codecov.io/github/wealdtech/walletd)
[![Go Report Card](https://goreportcard.com/badge/github.com/wealdtech/walletd)](https://goreportcard.com/report/github.com/wealdtech/walletd)

Daemon for accessing Ethereum 2 wallets and allowing protected signing operations to take place.

## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contribute](#contribute)
- [License](#license)

## Install

`walletd` is a standard Go module which can be installed with:

```sh
go get github.com/wealdtech/walletd
```

## Usage

`walletd` provies a gRPC interface to wallet operations such as listing accounts and signing requests.

### Configuration

The configuration file is at `config.json` and is held in the following location:

    - Windows: `%APPDATA%\wealdtech\walletd`
    - MacOSX: `${HOME}/Library/Application Support/wealdtech/walletd`
    - Linux: `${HOME}/.config/wealdtech/walletd`
  
A sample configuration file might look like:

```
{
  "stores": [
    {
      "name": "Local",
      "type": "filesystem"
    }
  ],
  "rules": [
    {
      "name": "Signer",
      "request": "sign",
      "account": "TODO",
      "script": "sign.lua"
    }
  ]
}
```

### Rules

Each time a signing request is sent to walletd it has the option to run a number of rules prior to signing the requested data.

  - request:
    - sign: general signing request
    - sign beacon proposal: sign a block proposal for the beacon chain
    - sign beacon attestation: sign a block attestation for the beacon chain
  - account:

### Scripts

Scripts are stored in the `scripts` directory of the default configuration location.  Scripts are written in the lua language.  A script must contain an `approve()` function that takes the parameters `request` and `storage`.  `request` contains information about the signing request, and `storage` contains persistent storage specific to the request and requesting account.

The `approve()` script should return one of the following three values:

  - `Approved` the signing can proceed
  - `Denied` the signing must not proceed
  - `Failed` the attempt to decide if the signing should go ahead or not has failed (which also implies that the signing must not proceed)

Multiple rules can match a single script.  In this situation all scripts are run one after the other, with a requirement for all scripts to return `Approved` before signing can proceed.

### Example

## Maintainers

Jim McDonald: [@mcdee](https://github.com/mcdee).

## Contribute

Contributions welcome. Please check out [the issues](https://github.com/wealdtech/walletd/issues).

## License

[Apache-2.0](LICENSE) © 2020 Weald Technology Trading Ltd
