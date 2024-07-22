FROM golang:1.22.2-bookworm

RUN apt update \
	&& apt -y upgrade

RUN apt -y install \
	libgmp-dev \
	build-essential \
	flex \
	bison

WORKDIR /usr/src/pbc-0.5.14
ADD ./build/pbc-0.5.14.tar.gz /usr/src
RUN ./configure \
	&& make \
	&& make install

RUN apt -y install \
	vim \
	jq

ARG GITHUB_TOKEN
ENV GITHUB_TOKEN ${GITHUB_TOKEN}
ENV WALLET_MNEMONIC test test test test test test test test test test test junk
ENV GOPRIVATE github.com/dpduado/dpduado-go

WORKDIR /app

RUN git config --global url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"

COPY ./app/go.mod /app/go.mod
COPY ./app/go.sum /app/go.sum

RUN go mod download

COPY --chmod=755 ./build/entrypoint.sh /entrypoint.sh
COPY ./build/config.json.template /etc/dpduado/harness/config.json.template

ENTRYPOINT ["/entrypoint.sh"]

