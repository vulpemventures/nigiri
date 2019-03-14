# bitcoin-box docker image
FROM ubuntu:18.04

# add bitcoind from the official PPA
# install bitcoind (from PPA) and make
RUN apt-get update && \
	apt-get install --yes wget clang cmake jq software-properties-common && \
	add-apt-repository --yes ppa:bitcoin/bitcoin && \
	apt-get update && \
	apt-get install --yes bitcoind

# install liquid
RUN mkdir -p /tmp && wget -qO- https://github.com/Blockstream/liquid/releases/download/liquid.3.14.1.22/liquid-3.14.1.22-x86_64-linux-gnu.tar.gz | tar xvz -C /tmp && \
  mv /tmp/liquid-3.14.1.22/bin/liquidd /usr/bin/liquidd && \ 
  mv /tmp/liquid-3.14.1.22/bin/liquid-cli /usr/bin/liquid-cli &&\
	mv /tmp/liquid-3.14.1.22/bin/liquid-tx /usr/bin/liquid-tx &&\
  rm -rf /tmp/liquid-3.14.1.22

RUN mkdir -p /config /script

# copy the box files into the image
ADD alice/config /config
ADD alice/run /script

WORKDIR /script
# expose two rpc ports for the nodes to allow outside container access
EXPOSE 19001 18884
STOPSIGNAL SIGINT

CMD ["./run"]
