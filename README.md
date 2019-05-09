# 🍣 Nigiri Bitcoin

Nigiri provides a selection of `docker-compose` batteries included to have ready-to-use bitcoin environment thats supports different networks and sidechains.

No time to make a Nigiri yourself?

```
$ curl getnigiri.vulpem.com | bash
``` 

## Utensils

* [Docker (compose)](https://docs.docker.com/compose/)

## Ingredients

* [Bitcoin daemon](https://bitcoin.org/en/bitcoin-core/)
* [Liquid daemon](https://blockstream.com/liquid/)
* [Electrum server](https://github.com/Blockstream/electrs)
* [Nigiri Chopsticks](https://github.com/vulpemventures/nigiri-chopsticks)

## Directions

| Preparation Time: 5 min  | Cooking Difficulty: Easy |
| --- | --- |

Clone the repo:

```bash
$ git clone https://github.com/vulpemventures/nigiri.git
```

Enter project directory and install dependencies:

```bash
$ bash scripts/install
```

This will create `~/.nigiri` copying there the `cli/resources/` directory.

Build binary (Mac version):
```
$ bash scripts/build darwin amd64
```

## Tasting

At the moment bitcoind, liquidd and electrs are started on *regtest* network. *testnet* and *mainnet* compose files will be released soon.


* Start nigiri:

```bash
$ nigiri start
```

Use the `--liquid` flag to let you do experiments with the Liquid sidechain. A liquid daemon and a block explorer
are also started when passing this flag.

* Stop nigiri:

```bash
$ nigiri stop
```

Use the `--delete` flag to not just stop Docker containers but also to remove them and delete the config file and any new data written in volumes.


Nigiri uses the default directory `~/.nigiri` to store configuration files and docker-compose files.
To set a custom directory use the `--datadir` flag.

## Nutrition Facts

The [list](https://github.com/blockstream/esplora/blob/master/API.md) of all available endpoints can be extended with one more `POST /faucet` which expects a body `{ "address": <receiving_address> }` by enabling faucet.

## Footnotes

If you really do love Sathoshi's favourite dish like us at Vulpem Ventures, check the real [recipe](https://www.allrecipes.com/recipe/228952/nigiri-sushi/) out and enjoy your own, delicious, hand made nigiri sushi.

## Roadmap

- [x] router
- [x] electrum server
- [x] bitcoin daemon
- [x] liquid daemon
- [x] block explorer UI
- [] regtest faucet
