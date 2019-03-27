FROM ubuntu:18.04

RUN apt-get update && apt-get install --yes wget

WORKDIR /build
RUN wget -qO- https://github.com/vulpemventures/electrs/releases/download/v0.4.1-bin/electrs.tar.gz | tar -xvz && rm -rf electrs.tar.gz

# RUN cargo build --release
WORKDIR /scripts
ADD electrs/run /scripts

# Electrum REST
EXPOSE 3002
STOPSIGNAL SIGINT

CMD ["./run"]