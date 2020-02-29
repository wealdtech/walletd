# walletd

[![Tag](https://img.shields.io/github/tag/wealdtech/walletd.svg)](https://github.com/wealdtech/walletd/releases/)
[![License](https://img.shields.io/github/license/wealdtech/walletd.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/wealdtech/walletd?status.svg)](https://godoc.org/github.com/wealdtech/walletd)
[![Travis CI](https://img.shields.io/travis/wealdtech/walletd.svg)](https://travis-ci.org/wealdtech/walletd)
[![codecov.io](https://img.shields.io/codecov/c/github/wealdtech/walletd.svg)](https://codecov.io/github/wealdtech/walletd)
[![Go Report Card](https://goreportcard.com/badge/github.com/wealdtech/walletd)](https://goreportcard.com/report/github.com/wealdtech/walletd)

Daemon holding Ethereum 2 keys and allowing signing operations to take place.

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

Configuration is held in the default location:

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
      "selector": ".*",
      "script": "signer.lua"
    }
  ]
}
```

Scripts are stored in the `scripts` directory of the default configuration location.

### Example

## Maintainers

Jim McDonald: [@mcdee](https://github.com/mcdee).

## Contribute

Contributions welcome. Please check out [the issues](https://github.com/wealdtech/walletd/issues).

## License

[Apache-2.0](LICENSE) © 2020 Weald Technology Trading Ltd
