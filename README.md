# 🍣 Nigiri Bitcoin 

Nigiri provides a command line interface that manages a selection of `docker-compose` batteries included to have ready-to-use bitcoin `regtest` development environment, with a **bitcoin** node, **electrum** explorer both backend and frontend user interface. 

It offers a [JSON HTTP proxy passtrough](https://github.com/vulpemventures/nigiri-chopsticks) that adds to the explorer handy endpoints like `/faucet` and automatic block generation when calling the `/tx` pushing a transaction.

You can have Elements too with the `--liquid` flag.


# No time to make a Nigiri yourself?
## Pre-built binary


* Download and install `nigiri` command line interface

```
$ curl https://getnigiri.vulpem.com | bash
```

This will create a directory `~/.nigiri` copying there `{bitcoin|elements}.conf` you can modify.


* Lauch Docker daemon (Mac OSX)

```
$ open -a Docker
``` 
You may want to [Manage Docker as a non-root user](https://docs.docker.com/engine/install/linux-postinstall/#manage-docker-as-a-non-root-user)

* Close and reopen your terminal, then start Bitcoin and Liquid

```
$ nigiri start --liquid
```
**That's it.**
Go to http://localhost:5000 for quickly inspect the Bitcoin blockchain or http://localhost:5001 for Liquid.

* Use the Bitcoin CLI inside the box

```
$ nigiri rpc getnewaddress "" "bech32"
bcrt1qsl4j5je4gu3ecjle8lckl3u8yywh8rff6xxk2r
```

* Use the Elements CLI inside the box

```
$ nigiri rpc --liquid getnewaddress "" "bech32"
el1qqwwx9gyrcrjrhgnrnjq9dq9t4hykmr6ela46ej63dnkdkcg8veadrvg5p0xg0zd6j3aug74cv9m4cf4jslwdqnha2w2nsg9x3
```

# Make from scratch
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
$ make install
```


* Build binary 

```
$ make build
```

Done! You should be able to find the binary in the local `./build` folder. Give it permission to execute and move/rename into your PATH.


* Clean

Remeber to always `clean` Nigiri before running `install` to upgrade to a new version.

```
$ make clean
```



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

* Generate and send bitcoin to given address

```bash
# Bitcoin
$ nigiri faucet bcrt1qsl4j5je4gu3ecjle8lckl3u8yywh8rff6xxk2r

# Elements
$ nigiri faucet --liquid el1qqwwx9gyrcrjrhgnrnjq9dq9t4hykmr6ela46ej63dnkdkcg8veadrvg5p0xg0zd6j3aug74cv9m4cf4jslwdqnha2w2nsg9x3
```

* Liquid only Issue and send a given quantity of an asset

```bash
$ nigiri mint el1qqwwx9gyrcrjrhgnrnjq9dq9t4hykmr6ela46ej63dnkdkcg8veadrvg5p0xg0zd6j3aug74cv9m4cf4jslwdqnha2w2nsg9x3 1000 VulpemToken VLP
```

* Broadcast a raw transaction and automatically generate a block

```bash
# Bitcoin
$ nigiri push <hex>

# Elements
$ nigiri push --liquid <hex>
```

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

* Use the Bitcoin CLI inside the box

```
$ nigiri rpc getnewaddress "" "bech32"
bcrt1qsl4j5je4gu3ecjle8lckl3u8yywh8rff6xxk2r
```

* Use the Elements CLI inside the box

```
$ nigiri rpc --liquid getnewaddress "" "bech32"
el1qqwwx9gyrcrjrhgnrnjq9dq9t4hykmr6ela46ej63dnkdkcg8veadrvg5p0xg0zd6j3aug74cv9m4cf4jslwdqnha2w2nsg9x3
```

* Run in headless mode (without Esplora)
If you are looking to spin-up Nigiri in Travis or Github Action you can use the `--ci` flag.

```
$ nigiri start --ci [--liquid]
```


* Update the docker images

```
$ nigiri update
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
