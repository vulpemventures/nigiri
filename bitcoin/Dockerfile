# bitcoin-box docker image
FROM ubuntu:18.04

# add bitcoind from the official PPA
# install bitcoind (from PPA) and make
RUN apt-get update && \
	apt-get install --yes clang cmake jq software-properties-common && \
	add-apt-repository --yes ppa:bitcoin/bitcoin && \
	apt-get update && \
	apt-get install --yes bitcoind

RUN mkdir -p /config /script

# copy the box files into the image
ADD bitcoin/config /config
ADD bitcoin/run /script

WORKDIR /script
# expose two rpc ports for the nodes to allow outside container access
EXPOSE 19001
STOPSIGNAL SIGINT

CMD ["./run"]
