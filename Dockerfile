FROM alpine:latest

RUN mkdir -p /opt/ringpop

COPY target/ringpop /bin/ringpop
COPY etc/hosts.json /opt/ringpop/hosts.json

ENTRYPOINT ["ringpop"]
