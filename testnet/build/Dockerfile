FROM debian:bullseye-slim as builder

ARG GETH_ALLTOOLS="geth-alltools-linux-amd64-1.14.3-ab48ba42"

RUN apt-get update && apt-get install -y \
    wget

RUN wget -O /tmp/${GETH_ALLTOOLS}.tar.gz https://gethstore.blob.core.windows.net/builds/${GETH_ALLTOOLS}.tar.gz
RUN tar zxf /tmp/${GETH_ALLTOOLS}.tar.gz -C /tmp
RUN mkdir -p /usr/local/geth/bin
RUN cp /tmp/${GETH_ALLTOOLS}/abigen   /usr/local/geth/bin/
RUN cp /tmp/${GETH_ALLTOOLS}/bootnode /usr/local/geth/bin/
RUN cp /tmp/${GETH_ALLTOOLS}/clef     /usr/local/geth/bin/
RUN cp /tmp/${GETH_ALLTOOLS}/evm      /usr/local/geth/bin/
RUN cp /tmp/${GETH_ALLTOOLS}/geth     /usr/local/geth/bin/
RUN cp /tmp/${GETH_ALLTOOLS}/rlpdump  /usr/local/geth/bin/

FROM ghcr.io/foundry-rs/foundry:latest@sha256:3bcbeab19b88d8a4245d811cf0d2cd35dbaa2042fd3f61516bae28156eedcd2a

COPY --from=builder /usr/local/geth/bin/* /usr/local/bin

RUN apk update \
    && apk upgrade \
    && apk add \
        vim \
        jq \
        yq

WORKDIR /app

COPY --chmod=755 ./entrypoint.sh /entrypoint.sh

# DO NOT EDIT!!! (0:SM, 1:SP, 2:TPA, 3:SU1, 4:SU2, 5:SU3)
ENV NUM_FIRST_ACCOUNTS 6
ENV RPC_HOST 0.0.0.0
ENV RPC_PORT 8545

ENTRYPOINT [ "/entrypoint.sh" ]
