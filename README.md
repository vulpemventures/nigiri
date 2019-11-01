# 🍣 Nigiri Bitcoin 

<img alt="Drone (self-hosted) with branch" src="https://img.shields.io/drone/build/vulpemventures/nigiri/master.svg?server=http%3A%2F%2Fdrone.vulpem.com">

Nigiri provides a selection of `docker-compose` batteries included to have ready-to-use bitcoin environment thats supports different networks and sidechains.



## Pre-built binary
No time to make a Nigiri yourself?


* Download and install `nigiri` command line interface

```
$ curl https://getnigiri.vulpem.com | bash
```

This will create a directory `~/.nigiri` copying there `{bitcoin|liquid}.conf` you can modify.


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

This will create `~/.nigiri` copying there the `{bitcoin|liquid}.conf` you can modify.

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

At the moment bitcoind, liquidd and electrs are started on *regtest* network. *testnet* and *mainnet* compose files will be released soon.


*  Start nigiri:

```bash
$ ./nigiri start
```

Use the `--liquid` flag to let you do experiments with the Liquid sidechain. A liquid daemon and a block explorer are also started when passing this flag.

* Stop nigiri:

```bash
$ ./nigiri stop
```

Use the `--delete` flag to not just stop Docker containers but also to remove them and delete the config file and any new data written in volumes.


Nigiri uses the default directory `~/.nigiri` to store configuration files and docker-compose files.
To set a custom directory use the `--datadir` flag.

Run the `help` command to see the full list of available flags.

## Nutrition Facts

The [list](https://github.com/blockstream/esplora/blob/master/API.md) of all available endpoints can be extended with one more `POST /faucet` which expects a body `{ "address": <receiving_address> }` by enabling faucet.

## Footnotes

If you really do love Sathoshi's favourite dish like us at Vulpem Ventures, check the real [recipe](https://www.allrecipes.com/recipe/228952/nigiri-sushi/) out and enjoy your own, delicious, hand made nigiri sushi.
