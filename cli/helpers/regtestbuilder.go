package helpers

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vulpemventures/nigiri/cli/builder"
)

func NewRegtestBuilder(rootPath string) builder.ComposeBuilder {
	rb := &regtestbuilder{}
	rb.New(rootPath)

	return rb
}

type regtestbuilder struct {
	path string
}

func (rb *regtestbuilder) New(rootPath string) {
	rb.path = rootPath
}

func (rb *regtestbuilder) Build() error {
	if err := buildComposeFile(rb.path); err != nil {
		return err
	}

	if err := buildBitcoinContainer(filepath.Join(rb.path, "bitcoin")); err != nil {
		return err
	}

	if err := buildElectrsContainer(filepath.Join(rb.path, "electrs")); err != nil {
		return err
	}

	return runCompose(rb.path)
}

func (rb *regtestbuilder) Delete() error {
	return deleteAll(rb.path)
}

func runCompose(path string) error {
	composePath := filepath.Join(path, "docker-compose.yml")
	cmd := exec.Command("docker-compose", "-f", composePath, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func buildComposeFile(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	r := strings.NewReplacer("\t", " ")
	if err := ioutil.WriteFile(filepath.Join(path, "docker-compose.yml"), []byte(r.Replace(composeFile)), os.ModePerm); err != nil {
		return err
	}

	return nil
}

func buildBitcoinContainer(path string) error {
	configPath := filepath.Join(path, "config")
	if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(path, "run"), []byte(btcRunFile), 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(path, "Dockerfile"), []byte(btcDockerFile), 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(configPath, "bitcoin.conf"), []byte(btcConfigFile), 0644); err != nil {
		return err
	}

	return nil
}

func buildElectrsContainer(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(path, "run"), []byte(electrsRunFile), 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(path, "Dockerfile"), []byte(electrsDockerFile), 0644); err != nil {
		return err
	}

	return nil
}

func deleteAll(path string) error {
	_, basename := filepath.Split(path)
	composePath := filepath.Join(path, "docker-compose.yml")
	volume := fmt.Sprintf("%s_bitcoin-config", basename)
	images := []string{
		fmt.Sprintf("%s_bitcoin", basename),
		fmt.Sprintf("%s_electrs", basename),
	}

	cmdCompose := fmt.Sprintf("docker-compose -f %s down", composePath)
	cmdVolume := fmt.Sprintf("docker volume rm %s", volume)
	cmdImage := fmt.Sprintf("docker rmi %s", strings.Join(images, " "))
	cmdDelete := fmt.Sprintf("rm -rf %s", path)

	cmd := exec.Command(
		"/bin/sh", "-c",
		fmt.Sprintf("%s; %s; %s; %s", cmdCompose, cmdVolume, cmdImage, cmdDelete),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

const composeFile = `version: '3'
services:
	bitcoin:
		build: 
			context: bitcoin/
			dockerfile: Dockerfile
		networks:
			local:
				ipv4_address: 10.10.0.10
		volumes:
			- bitcoin-config:/config
	electrs:
		build:
			context: electrs/
			dockerfile: Dockerfile
		networks:
			local:
				ipv4_address: 10.10.0.12
		volumes:
			- bitcoin-config:/config
	# chopsticks:
	# 	build:
	# 		context: chopsticks/
	#		dockerfile: Dockerfile
	# 	ports:
	# 		- 3000:3000
	# 	networks:
	# 		local:
	# 			ipv4_address: 10.10.0.13

networks:
	local:
		driver: bridge
		ipam:
			config:
				- subnet: 10.10.0.0/24

volumes:
	bitcoin-config:`

const btcDockerFile = `FROM ubuntu:18.04

RUN apt-get update && \
	apt-get install --yes clang cmake jq software-properties-common && \
	add-apt-repository --yes ppa:bitcoin/bitcoin && \
	apt-get update && \
	apt-get install --yes bitcoind

RUN mkdir -p /config /script

ADD config /config
ADD run /script

WORKDIR /script

EXPOSE 19001
STOPSIGNAL SIGINT

CMD ["./run"]`

const btcRunFile = `#!/bin/bash
set -ex

b1="bitcoin-cli -datadir=/config"

function clean {
	$b1 stop
}

trap clean SIGINT

bitcoind -datadir=/config &

sleep 10

$b1 generate 200
wait $!`

const btcConfigFile = `regtest=1
testnet=0
dnsseed=0
upnp=0

[regtest]
port=19000
rpcport=19001

server=1
txindex=0

rpcuser=admin1
rpcpassword=123
rpcallowip=0.0.0.0/0`

const electrsDockerFile = `FROM ubuntu:18.04

RUN apt-get update && apt-get install --yes wget

WORKDIR /build
RUN wget -qO- https://github.com/vulpemventures/electrs/releases/download/v0.4.1-bin/electrs.tar.gz | tar -xvz && rm -rf electrs.tar.gz

WORKDIR /scripts
ADD run /scripts

EXPOSE 3002
STOPSIGNAL SIGINT

CMD ["./run"]`

const electrsRunFile = `#!/bin/bash
set -e

function clean {
  kill -9 $(pidof electrs)
}
trap clean SIGINT

/build/electrs -vvvv --network regtest --daemon-dir /config --daemon-rpc-addr="10.10.0.10:19001" --cookie="admin1:123" --http-addr="0.0.0.0:3002" &
wait $!`
