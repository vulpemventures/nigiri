FROM ubuntu:18.04

RUN apt-get update && apt-get install --yes wget

RUN mkdir -p /tmp && wget -qO- https://github.com/Blockstream/liquid/releases/download/liquid.3.14.1.22/liquid-3.14.1.22-x86_64-linux-gnu.tar.gz | tar xvz -C /tmp && \
  mv /tmp/liquid-3.14.1.22/bin/liquidd /usr/bin/liquidd && \ 
  mv /tmp/liquid-3.14.1.22/bin/liquid-cli /usr/bin/liquid-cli &&\
	mv /tmp/liquid-3.14.1.22/bin/liquid-tx /usr/bin/liquid-tx &&\
  rm -rf /tmp/liquid-3.14.1.22

RUN mkdir -p /config /script

ADD liquid/config /config
ADD liquid/run /script

WORKDIR /script

EXPOSE 18884
STOPSIGNAL SIGINT

CMD ["./run"]
