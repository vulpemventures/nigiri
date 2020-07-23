# 🍣 Nigiri Bitcoin 

Nigiri provides a command line interface that manages a selection of `docker-compose` batteries included to have ready-to-use bitcoin `regtest` development environment, with a **bitcoin** node, **electrum** explorer both backend and frontend user interface. 

It offers a [JSON HTTP proxy passtrough](https://github.com/vulpemventures/nigiri-chopsticks) that adds to the explorer handy endpoints like `/faucet` and automatic block generation when calling the `/tx` pushing a transaction.

You can have Elements too with the `--liquid` flag.



## Pre-built binary
No time to make a Nigiri yourself?


* Download and install `nigiri` command line interface

```
$ curl https://getnigiri.vulpem.com | bash
```

This will create a directory `~/.nigiri` copying there `{bitcoin|elements}.conf` you can modify.


* Lauch Docker daemon (Mac OSX)

```
$ open -a Docker
``` 

* Close and reopen your terminal, then start Bitcoin and Liquid

```
$ nigiri start --liquid
```
That's it.
Go to http://localhost:5000 for quickly inspect the Bitcoin blockchain or http://localhost:5001 for Liquid.

## Utensils

* [Docker (compose)](https://docs.docker.com/compose/)
* Go

## Ingredients

* [Bitcoin daemon](https://bitcoin.org/en/bitcoin-core/)
* [Liquid daemon](https://blockstream.com/liquid/)
* [Electrum server](https://github.com/Blockstream/electrs)
* [Esplora](https://github.com/Blockstream/esplora)
* [Nigiri Chopsticks](https://github.com/vulpemventures/nigiri-chopsticks)

## Directions

| Preparation Time: 5 min  | Cooking Difficulty: Easy |
| --- | --- |

* Clone the repo:

```bash
$ git clone https://github.com/vulpemventures/nigiri.git
```

* Enter project directory and install dependencies:

```bash
$ bash scripts/install
```

This will create `~/.nigiri` copying there the `{bitcoin|elements}.conf` you can modify.

* Build binary 
```
# MacOSX
$ bash scripts/build darwin amd64
# Linux 
$ bash scripts/build linux amd64
```

* Remove nigiri
```
$ bash scripts/clean
```

Note: Remeber to always `clean` Nigiri before running `install` after a pull.

## Tasting

At the moment bitcoind, elements and electrs are started on *regtest* network.


*  Start nigiri:

```bash
$ nigiri start
```
Use the `--liquid` flag to let you do experiments with the Liquid sidechain. A liquid daemon and a block explorer are also started when passing this flag.

* Stop nigiri:

```bash
$ nigiri stop
```
Use the `--delete` flag to not just stop Docker containers but also to remove them and delete the config file and any new data written in volumes.


* Check the logs of Bitcoin services

```bash
# Bitcoind
$ nigiri logs node
# Electrs
$ nigiri logs electrs
# Chopsticks
$ nigiri logs chopsticks
```

* Check the logs of Liquid services

```bash
# Elementsd
$ nigiri logs node --liquid
# Electrs Liquid
$ nigiri logs electrs --liquid
# Chopsticks Liquid
$ nigiri logs chopsticks --liquid
```

Nigiri uses the default directory `~/.nigiri` to store configuration files and docker-compose files.
To set a custom directory use the `--datadir` flag.

Run the `help` command to see the full list of available flags.

## Nutrition Facts

`Chopsticks` service exposes on port `3000` (and on `3001` if started with `--liquid` flag) all [Esplora's available endpoints](https://github.com/blockstream/esplora/blob/master/API.md) and extends them with the following:


### Bitcoin & Liquid

- `POST /faucet` which expects a body `{ "address": <receiving_address> }` 
- `POST /tx` has been extended to automatically mine a block when is called.

### Liquid only

- `POST /mint` which expects a body `{"address": "ert1q90dz89u8eudeswzynl3p2jke564ejc2cnfcwuq", "quantity": 1000, "name":"VULPEM", "ticker":"VLP"}` 
- `POST /registry` to get extra info about one or more assets like `name` and `ticker` which expects a body with an array of assets `{"assets": ["2dcf5a8834645654911964ec3602426fd3b9b4017554d3f9c19403e7fc1411d3"]}`


## Footnotes

If you really do love Sathoshi's favourite dish like us at Vulpem Ventures, check the real [recipe](https://www.allrecipes.com/recipe/228952/nigiri-sushi/) out and enjoy your own, delicious, hand made nigiri sushi.
