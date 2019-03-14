FROM rust:latest

RUN apt-get update
RUN apt-get install -y clang cmake

WORKDIR /build
RUN wget -qO- https://github.com/vulpemventures/electrs/releases/download/v0.4.1-bin/electrs.tar.gz | tar -xvz && rm -rf electrs.tar.gz
RUN ls

# RUN git clone https://github.com/blockstream/electrs
# WORKDIR /home/user/electrs
# RUN git checkout new-index

# RUN cargo build --release
WORKDIR /scripts
ADD electrs/run /scripts

# Electrum REST
EXPOSE 3002
STOPSIGNAL SIGINT

CMD ["./run"]