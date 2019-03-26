# 🍣 Nigiri Bitcoin

A dockerized environment hosting a bitcoin and liquid daemons in regtest network with an electrum server that indexes and keeps track of all UTXOs.

## Utensils

* [Docker (compose)](https://docs.docker.com/compose/)

## Ingredients

* [Bitcoin daemon](https://bitcoin.org/en/bitcoin-core/)
* [Liquid daemon](https://blockstream.com/liquid/)
* [Electrum server](https://github.com/Blockstream/electrs)
* [Chopsticks](https://github.com/vulpemventures/nigiri-chopsticks)

## Directions

| Preparation Time: 20 min  | Cooking Difficulty: Easy |
| --- | --- |

Clone the repo:

```bash
$ git clone https://github.com/vulpemventures/nigiri.git && cd nigiri
```

Create and start nigiri (only the first time or after cleaning):

```bash
$ bash scripts/create
```

At the moment bitcoind, liquidd and electrs are started on *regtest* network.

Start nigiri:

```bash
$ bash scripts/start
```

This will start 4 containers that run the following services respectevely:

* bitcoin daemon (regtest)
* liquid daemon
* electrs REST server
* API passthrough with optional faucet and mining capabilities (nigiri-chopsticks)

Stop nigiri:

```bash
$ bash scripts/stop
```

Stop and uninstall nigiri:

```bash
$ bash scripts/clean
```

When setup is completed, you can call any endpoint at `http://localhost:3000/`.

## Nutrition Facts

The [list](https://github.com/blockstream/esplora/blob/master/API.md) of all available endpoints has been extended with one more `POST /send` which expects a body `{ "address": <receiving_address> }`

## Footnotes

If you really do love Sathoshi's favourite dish like us at Vulpem Ventures, check the real [recipe](https://www.allrecipes.com/recipe/228952/nigiri-sushi/) out and enjoy your own, delicious, hand made nigiri sushi.

## Roadmap

- [x] router
- [x] electrum server
- [x] bitcoin daemon
- [x] liquid daemon
- [x] faucet
