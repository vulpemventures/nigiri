# 🍣 Nigiri Bitcoin

Nigiri provides a command line interface that manages a selection of `docker-compose` batteries included to have a ready-to-use Bitcoin `regtest` development environment. Out of the box, you get:

- **Bitcoin Node**: A Bitcoin Core node running in regtest mode
- **Electrum**: Backend and frontend explorer for quick blockchain inspection
- **Chopsticks**: A [JSON HTTP proxy](https://github.com/vulpemventures/nigiri-chopsticks) that adds handy endpoints like `/faucet` and automatic block generation

You can extend your setup with:

- **Ark**: A Bitcoin layer two implementation for scalable off-chain transactions
- **Elements/Liquid sidechain** with `--liquid` flag
- **Lightning Network nodes** with `--ln` flag (Core Lightning, LND, and Taproot Assets)

# No time to make a Nigiri yourself?

## Pre-built binary

- Download and install `nigiri` command line interface

```bash
curl https://getnigiri.vulpem.com | bash
```

> [!NOTE]
> On Windows, you are heavily encouraged to make use of [WSL](https://learn.microsoft.com/en-us/windows/wsl). Ensure that Docker is configured to integrate and use WSL. Prepending most commands in this document with `wsl` is sufficient.

This will also install several configurable files, such as `bitcoin.conf` and `elements.conf`, that can be edited. These can be found browsing the following directory:

POSIX (Linux/BSD): `~/.nigiri`

macOS: `$HOME/Library/Application\ Support/Nigiri`

Windows (WSL): `~/.nigiri`

Windows: `%LOCALAPPDATA%\Nigiri`

Plan 9: `$home/nigiri`

- Launch Docker daemon (Mac OSX)

```bash
open -a Docker
```

You may want to [Manage Docker as a non-root user](https://docs.docker.com/engine/install/linux-postinstall/#manage-docker-as-a-non-root-user)

- Close and reopen your terminal, then start Bitcoin and Ark

```bash
nigiri start --ark
```

**That's it.**

Go to http://localhost:5000 for quickly inspect the Bitcoin blockchain.

Want more? Add Elements/Liquid, Lightning nodes, or Ark:

```bash
nigiri start --ark --liquid  # Add Elements/Liquid sidechain
nigiri start --ark --ln      # Add Lightning Network nodes
nigiri start --liquid --ln   # Add both Liquid and Lightning
nigiri start --ark --liquid --ln  # Add all features
```

**Note for users of macOS Monterey an onward**

<details>
  <summary>Show more...</summary>
   When trying to start Nigiri, you might get an error similar to the following:

```bash
Error response from daemon: Ports are not available: listen tcp 0.0.0.0:5000: bind: address already in use
exit status 1
```

This is due to AirPlay Receiver using port 5000, conflicting with Esplora trying to run using the very same port.

There are two ways to deal with this issue:

1. Uncheck AirPlay Receiver in `System Preferences → Sharing → AirPlay Receiver`
2. Change Esplora’s port to something other than 5000. This can be done by changing it in [docker-compose.yml](https://github.com/vulpemventures/nigiri/blob/master/cmd/nigiri/resources/docker-compose.yml#L110) found in your data directory. If you previously tried starting Nigiri getting an error – you might have to run `nigiri stop --delete` before restarting it.
</details>
<br />

## Tasting

At the moment bitcoind, elements and electrs are started on _regtest_ network.

### Start nigiri

```bash
nigiri start
```

- Use the `--liquid` flag to let you do experiments with the Liquid sidechain. A liquid daemon and a block explorer are also started when passing this flag.

- Use the `--ln` flag to start a Core Lightning node, a LND node and a Tap daemon.

- Use the `--remember` flag to save the currently used flags (like `--liquid`, `--ln`, `--ark`) so they are automatically applied on subsequent `nigiri start` calls without needing to specify them again.

### Stop nigiri

```bash
nigiri stop
```

Use the `--delete` flag to not just stop Docker containers but also to remove them and delete the config file and any new data written in volumes.

### Forget remembered flags

```bash
nigiri forget
```

This command removes any flags previously saved using `nigiri start --remember`, causing future `nigiri start` calls to use default flags unless specified.

### Generate and send bitcoin to given address

```bash
# Bitcoin
nigiri faucet <bitcoin_address>
## Fund the Core Lightning node
nigiri faucet cln 0.01

## Fund the LND node
nigiri faucet lnd 0.01

# Elements
nigiri faucet --liquid <liquid_address>
```

### Send Liquid asset to given address

```bash
nigiri faucet --liquid <liquid_address> <amt> <liquid_asset>
```

### **Liquid only** Issue and send a given quantity of an asset

```bash
nigiri mint <liquid_address> 1000 VulpemToken VLP
```

### Broadcast a raw transaction and automatically generate a block

```bash
# Bitcoin
nigiri push <hex>

# Elements
nigiri push --liquid <hex>
```

### Check the logs of Bitcoin services

```bash
# Bitcoind
nigiri logs bitcoin
# Electrs
nigiri logs electrs
# Chopsticks
nigiri logs chopsticks
```

### Check the logs of Liquid services

```bash
# Elementsd
nigiri logs liquid
# Electrs Liquid
nigiri logs electrs-liquid
# Chopsticks Liquid
nigiri logs chopsticks-liquid
```

### Check the logs of Lightning services

```bash
# Core Lightning
nigiri logs cln
# LND
nigiri logs lnd
```

### Use the Bitcoin CLI inside the box

```bash
nigiri rpc getnewaddress "" "bech32"
bcrt1qsl4j5je4gu3ecjle8lckl3u8yywh8rff6xxk2r
```

### Use the Elements CLI inside the box

```bash
nigiri rpc --liquid getnewaddress "" "bech32"
el1qqwwx9gyrcrjrhgnrnjq9dq9t4hykmr6ela46ej63dnkdkcg8veadrvg5p0xg0zd6j3aug74cv9m4cf4jslwdqnha2w2nsg9x3
```

### Use the Core Lightning & LND CLIs inside the box

```bash
# Core Lightning
nigiri cln listpeers
# LND
nigiri lnd listpeers
# Tap (Taro)
nigiri tap assets list
```

### Connect Core Lightning to LND

```bash
nigiri cln connect `nigiri lnd getinfo | jq -r .identity_pubkey`@lnd:9735
```

### Open a channel between cln and lnd

```bash
nigiri lnd openchannel --node_key=`nigiri cln getinfo | jq -r .id` --local_amt=100000
nigiri cln fundchannel `nigiri lnd getinfo | jq -r .identity_pubkey` 100000
```

### Pay invoices between cln and lnd

```bash
nigiri lnd payinvoice `nigiri cln invoice 21000 $(date +%s) "test" | jq -r .bolt11`
nigiri cln pay `nigiri lnd addinvoice 21 | jq -r .payment_request`
```

### Use the Ark CLI inside the box

```bash
# Check versions
nigiri ark --version    # or -v
nigiri arkd --version   # or -v


# Use the Ark client
nigiri ark config                # Show wallet configuration
nigiri ark receive              # Show receiving addresses
nigiri ark balance              # Show wallet balance

# Configure the Ark daemon client
nigiri arkd wallet status       # Show wallet status
nigiri arkd wallet create --password secret  # Create a new wallet
nigiri arkd wallet unlock --password secret  # Unlock the wallet
nigiri arkd wallet status       # Verify ark wallet is set up

# Initialize the Ark client (only needed once)
nigiri ark init --network regtest --password secret --server-url localhost:7070 --explorer http://chopsticks:3000
```

### Update the docker images

```
nigiri update
```

Nigiri uses the default directory `~/.nigiri` to store configuration files and docker-compose files.
To set a custom directory use the `--datadir` flag.

Run the `help` command to see the full list of available flags.

# Make from scratch

## Utensils

- [Docker (compose)](https://docs.docker.com/compose/)
- Go

## Ingredients

- [Bitcoin daemon](https://bitcoin.org/en/bitcoin-core/)
- [Liquid daemon](https://blockstream.com/liquid/)
- [Electrum server](https://github.com/Blockstream/electrs)
- [Esplora](https://github.com/Blockstream/esplora)
- [Nigiri Chopsticks](https://github.com/vulpemventures/nigiri-chopsticks)

## Directions

| Preparation Time: 5 min | Cooking Difficulty: Easy |
| ----------------------- | ------------------------ |

- Clone the repo:

```bash
git clone https://github.com/vulpemventures/nigiri.git
```

- Enter project directory and install dependencies:

```bash
make install
```

- Build binary

```
make build
```

Done! You should be able to find the binary in the local `./build` folder. Give it permission to execute and move/rename into your PATH.

- Clean

Remember to always `clean` Nigiri before running `install` to upgrade to a new version.

```bash
make clean
```

## Nutrition Facts

`Chopsticks` service exposes on port `3000` (and on `3001` if started with `--liquid` flag) all [Esplora's available endpoints](https://github.com/blockstream/esplora/blob/master/API.md) and extends them with the following:

### Bitcoin & Liquid

- `POST /faucet` which expects a body `{ "address": <receiving_address> }`
- `POST /tx` has been extended to automatically mine a block when is called.

### Liquid only

- `POST /mint` which expects a body `{"address": "ert1q90dz89u8eudeswzynl3p2jke564ejc2cnfcwuq", "quantity": 1000, "name":"VULPEM", "ticker":"VLP"}`
- `POST /registry` to get extra info about one or more assets like `name` and `ticker` which expects a body with an array of assets `{"assets": ["2dcf5a8834645654911964ec3602426fd3b9b4017554d3f9c19403e7fc1411d3"]}`

## Footnotes

If you really do love Satoshi's favourite dish like us at Vulpem Ventures, check the real [recipe](https://www.allrecipes.com/recipe/228952/nigiri-sushi/) out and enjoy your own, delicious, hand made nigiri sushi.
