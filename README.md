# 🍣 Nigiri Bitcoin

Nigiri provides a fully dockerized ready-to-use bitcoin environment thats supports different networks and chains.

## Utensils

* [Docker (compose)](https://docs.docker.com/compose/)

## Ingredients

* [Bitcoin daemon](https://bitcoin.org/en/bitcoin-core/)
* [Liquid daemon](https://blockstream.com/liquid/)
* [Electrum server](https://github.com/Blockstream/electrs)
* [Chopsticks](https://github.com/vulpemventures/nigiri-chopsticks)

## Directions

| Preparation Time: 5 min  | Cooking Difficulty: Easy |
| --- | --- |

Clone the repo:

```bash
$ git clone https://github.com/vulpemventures/nigiri.git
```

Enter project directory and install dependencies:

```bash
nigiri $ bash scripts/install
```

This will create `~/.nigiri` copying there the `cli/resources/` directory.

Build binary (Mac version):
```
nigiri $ bash scripts/build darwin amd64
```

At the moment bitcoind, liquidd and electrs are started on *regtest* network.

Start nigiri:

```bash
nigiri/build $ nigiri-linux-amd64 start
```

Nigiri uses the default directory `~/.nigiri` to store the configuration file and docker stuff.
To set a custom directory use the `--datadir` flag, but do not forget to always pass this flag to other commands, just as you do with your `bitcoind`.  

The environment will start with 3 containers for `regtest` bitcoin network that run the following services respectevely:

* bitcoin daemon
* electrs REST server
* API passthrough with optional faucet and mining capabilities (nigiri-chopsticks)

Use the `--liquid` flag to let you do experiments with the Liquid sidechain. A liquid daemon and a block explorer
are also started when passing this flag.

Stop nigiri:

```bash
nigiri/build $ nigiri-linux-amd64 stop
```

Use the `--delete` flag to not just stop Docker containers but also to remove them and delete the config file and any new data written in volumes.

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
