FROM ubuntu:18.04

RUN apt-get update && apt-get install --yes curl wget

RUN mkdir -p /tmp /config /script

ADD liquid/config /config
ADD liquid/script /script

WORKDIR /script

RUN /script/install

EXPOSE 18884
STOPSIGNAL SIGINT

CMD ["./run"]
