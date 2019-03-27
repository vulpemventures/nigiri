FROM ubuntu:18.04

WORKDIR /build
COPY chopsticks/nigiri-chopsticks-linux-amd64.tar.gz /build
RUN ls
RUN tar -xzf nigiri-chopsticks-linux-amd64.tar.gz && rm -rf *.tar.gz

EXPOSE 3000

CMD ["./nigiri-chopsticks-linux-amd64", "--addr", "0.0.0.0:3000", "--electrs-addr", "10.10.0.12:3002", "--rpc-addr", "10.10.0.10:19001", "--use-logger", "true"]