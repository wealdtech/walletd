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

`walletd` provies a gRPC interface to wallet operations such as listing accounts and signing requests.  The daemon provides a number of security measures to avoid unauthorised uses of the private keys, and protection against invalid actions (_e.g._ slashing events).

## Architecture

### Configuration directory and files

The default configuration directory is at `config.json` and is held in the following location:

    - Windows: `%APPDATA%\wealdtech\walletd`
    - MacOSX: `${HOME}/Library/Application Support/wealdtech/walletd`
    - Linux: `${HOME}/.config/wealdtech/walletd`
  
This will usually contain the following files:

  - `config.json` the overall configuration file for `walletd`
  - `perms.json` permissions for each client certificate
  - `security` a directory containing certificates for the server and client certificate authority

These items are explained in more detail below.

### Example

The architecture we want to achieve is shown below:

![Validator architecture](images/architecture.png)

In this architecture we have three validators clients.  Validator clients 1 and 2 are in a cluster, and between them manage accounts 1, 2, and 3.  Validator client 3 is standalone, and manages account 4.

#### Creating wallets and accounts
The first step is to create some wallets and validator keys for said wallets, using [ethdo](https://github.com/wealdtech/ethdo):

```
$ ethdo wallet create --wallet=wallet1
$ ethdo account create --account=wallet1/account1 --passphrase=secret
$ ethdo account create --account=wallet1/account2 --passphrase=secret
$ ethdo account create --account=wallet1/account3 --passphrase=secret
$ ethdo wallet create --wallet=wallet2
$ ethdo account create --account=wallet2/account4 --passphrase=secret
```

Here we have two wallets, one for each set of validator clients.  It is possible for different wallets to have different features, such as level of security and location, but for the purposes of this example they are both standard (non-deterministic) wallets (see ethdo documentation for other options).

#### Creating certificates
We need a certificate for the wallet daemon.  We could use a certificate from a well-known certificate authority such as LetsEncrypt, or we could create our own; we will create our own using [certstrap](https://github.com/square/certstrap).

First, we create the certificate authority.  Note the key created in this process is critical to the security of your deposits and should be protected with all reasonable measures; this should include a passphrase when promted.
```
$ certstrap --depot-path . init --common-name "Wallet daemon authority" --expires "3 years"
Enter passphrase (empty for no passphrase): 
Enter same passphrase again: 
Created ./Wallet_daemon_authority.key (encrypted by passphrase)
Created ./Wallet_daemon_authority.crt
Created ./Wallet_daemon_authority.crl
```

The server needs its own certificate.  We use the sample name `server.example.com` here but you should replace this with the name of your server.  If you are testing `walletd` locally you can use `localhost` instead of the server name.
```
$ certstrap --depot-path . request-cert --common-name server.example.com
Enter passphrase (empty for no passphrase): 
Enter same passphrase again: 
Created ./server.example.com.key
Created ./server.example.com.csr
$ certstrap --depot-path . sign --CA "Wallet daemon authority" --expires="3 years" server.example.com
Enter passphrase for CA key (empty for no passphrase): 
Created ./server.example.com.crt from ./server.example.com.csr signed by ./Wallet_daemon_authority.key
```

Next, we create and sign certificates for the three clients that will be connecting to the daemon.  Note the keys created here should not have a passphrase supplied; they will reside with the valdiator clients so use of the key is should be possible without requiring human intervention (to allow for server restarts _etc._).  For the first client:

```
$ certstrap --depot-path . request-cert --common-name client1
Enter passphrase (empty for no passphrase): 
Enter same passphrase again: 
Created ./client1.key
Created ./client1.csr
$ certstrap --depot-path . sign --CA "Wallet daemon authority" --expires="3 years" client1
Enter passphrase for CA key (empty for no passphrase): 
Created ./client1.crt from ./client1.csr signed by ./Wallet_daemon_authority.key
```

and the same commands can be used for the other clients, using "client2" and "client3" in place of "client1".  At this point you should have the following files:

  - `client1.crt`: the signed certificate for client1; needs to be moved to the server running client1
  - `client1.csr`: the signing request for client1; can be deleted
  - `client1.key`: the key for client1; needs to be moved to the server running client1
  - `client2.crt`: the signed certificate for client2; needs to be moved to the server running client3
  - `client2.csr`: the signing request for client2; can be deleted
  - `client2.key`: the key for client2; needs to be moved to the server running client3
  - `client3.crt`: the signed certificate for client3; needs to be moved to the server running client3
  - `client3.csr`: the signing request for client3; can be deleted
  - `client3.key`: the key for client3; needs to be moved to the server running client3
  - `server.example.com.crt`: the certificate for `walletd`; needs to be moved to the server running the `walletd`
  - `server.example.com.csr`: the signing request for `walletd`; can be deleted
  - `server.example.com.key`: the key for `walletd`; needs to be moved to the server running the `walletd`
  - `Wallet_daemon_authority.crl`: the certificate revocation list for the wallet daemon; needs to be copied to the server running the wallet daemon
  - `Wallet_daemon_authority.crt`: the certificate for the wallet daemon; needs to be copied to all servers running clients
  - `Wallet_daemon_authority.key`: the key for the wallet daemon; needs to be copied to the server running the wallet daemon

To provide the certificates for the wallet daemon make a directory `security` in the configuration directory as defined above and copy the `server.example.com.crt` and `server.example.com.key` files in to it.  Also copy `Wallet_daemon_authority.crt` to the same directory with the name `ca.crt`.  The contents of the `security` directory in your configuration directory should be:

  - `ca.crt`: copy of `Wallet_daemon_authority.crt` from the previous step
  - `server.example.com.crt`: copy of `server.example.com.crt` from the previous step
  - `server.example.com.key`: copy of `server.example.com.key` from the previous step

At this point you also need a minimal `config.json` file so `walletd` knows which certificates to use.  You can create this in the configuration directory stated above with the contents:

```json
{
  "server": {
    "name": "server.example.com"
  }
}
```

You can check the configuration of the certificates by running the command:

```sh
$ walletd --show-certs
Server certificate issued by: Wallet daemon authority
Server certificate expires: 2023-03-24 13:47:19 +0000 UTC
Server certificate issued to: server.example.com

Certificate authority certificate is: Wallet daemon authority
Certificate authority certificate expires: 2023-03-24 13:47:20 +0000 UTC
```

#### Mapping certificates to keys
The next step is to configure `walletd` to know which clients have access to which keys.  This is defined in the `perms.json` file, which should reside in the same directory as `config.json`.  For our purposes we need:

```
"certificates": [
  { 
    "name": "client1",
    "permissions": [
      {
        "path": "wallet1",
        "operations": ["All"]
      }
    ]
  },
  {
    "name": "client2",
    "permissions": [
      {
        "path": "wallet1",
        "operations": ["All"]
      }
    ]
  },
  {
    "name": "client3",
    "permissions": [
      {
        "path": "wallet2",
        "operations": ["All"]
      }
    ]
  }
]
```

Once this is in place it can be confirmed by running `walletd --show-perms`:

```
$ walletd --show-perms
Permissions for "client1":
	- accounts matching the path "wallet1" can carry out all operations
Permissions for "client2":
	- accounts matching the path "wallet1" can carry out all operations
Permissions for "client3":
	- accounts matching the path "wallet2" can carry out all operations
```

#### Starting `walletd`

To start `walletd` type:

```sh
$ walletd
WARN[0000] No stores configured; using default          
badger 2020/03/26 15:22:20 INFO: All 0 tables opened in 0s
INFO[0000] Listening                                     address=":12346"
```

`walletd` will provide information about requests it receives so this window should be monitored for errors.

#### Testing client certificates
`ethdo` interacts with the wallet daemon using the `--remote` `--client-cert`  and `--client-key` options.  For example, to list accounts accessible in `wallet1` with the `client1` certificate:

```sh
$ ethdo --remote=server.example.com:12346 --client-cert=client1.crt --client-key=client1.key --server-ca-cert=Wallet_daemon_authority.crt wallet accounts --wallet=wallet1
account1
account3
account2
```

As would be expected from the configured permissions, `client3` cannot access the accounts in `wallet1`:

```sh
$ ethdo --remote=server.example.com:12346 --client-cert=client3.crt --client-key=client3.key --server-ca-cert=Wallet_daemon_authority.crt wallet accounts --wallet=wallet1
```

At this point it has been confirmed that the client certificates operate as expected, and that walletd is appropriately configured.  The client certificates can now be used by validators to remotely access their keys.

## Custom rules

`walletd` comes with a rules engine that allows users to create their own set of conditions under which actions can take place (or not).  Whenever a request is sent to `walletd` it runs rules based on the request and account carrying out the request.

## Writing rule scripts

Rule scripts are written in the lua language.  A script must contain an `approve()` function that takes the following parameters:

  - `request`: a table with request-specific information.  For example, a signing request will have information about the data to be signed and its signing domain.
  - `storage`: a table with access to persistent storage.  The storage is specific to this (request type, account) tuple.  All data in this table will be written to persistent storage on completion of the script (regardless of whether it results in an approval or denial, however not on failure)
  - `messages`: a table which starts empty.  All data in this table will be written to the `walletd` log file on completion of the script (regardless of whether it results in an approval or denial, however not on failure)

The `approve()` script should return one of the following three values:

  - `Approved` the signing can proceed
  - `Denied` the signing must not proceed
  - `Failed` the attempt to decide if the signing should go ahead or not has failed (which also implies that the signing must not proceed)

To provide an example: a validator should only sign a single beacon block proposal for a given slot, so if there is more than one attempt to sign a request for a given slot it should be denied.  A script to carry this out may look like the following:

```lua
function approve(request, storage, messages)
  if storage.slot ~= nil and storage.slot <= request.slot then
    table.insert(messages, string.format("Request slot %d equal to or lower than previous signed slot %s", request.slot, storage.slot))
    return "Denied"
  end
  storage.slot = request.slot
  return "Approved"
end
```

This ensures that any attempt to sign a beacon block proposal whose slot is equal to or lower than a previously successful signature will be denied.

## Configuring rules

Rule information is configured in the `config.json` file under a `rules` entry.
Multiple rules can match a single script.  In this situation all scripts are run one after the other, with a requirement for all scripts to return `Approved` before signing can proceed.

A sample `config.json` with rules may look like the below:

```json
{
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

### Stores

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

## Maintainers

Jim McDonald: [@mcdee](https://github.com/mcdee).

## Contribute

Contributions welcome. Please check out [the issues](https://github.com/wealdtech/walletd/issues).

## License

[Apache-2.0](LICENSE) © 2020 Weald Technology Trading Ltd
