FROM ghcr.io/foundry-rs/foundry:latest@sha256:3bcbeab19b88d8a4245d811cf0d2cd35dbaa2042fd3f61516bae28156eedcd2a

RUN apk update \
    && apk upgrade \
    && apk add \
        vim \
        jq \
        yq

WORKDIR /app

COPY --chmod=755 ./asset/entrypoint.sh /entrypoint.sh

# DO NOT EDIT!!! (0:SM, 1:SP, 2:TPA, 3:SU1, 4:SU2, 5:SU3)
ENV NUM_FIRST_ACCOUNTS 6
ENV RPC_HOST 0.0.0.0
ENV RPC_PORT 8545

ENTRYPOINT [ "/entrypoint.sh" ]
