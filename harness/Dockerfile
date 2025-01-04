FROM prism/prism-go

RUN apt update \
	&& apt -y install \
		libgmp-dev \
		build-essential \
		flex \
		bison \
		vim \
		jq

WORKDIR /app

COPY ./app/go.mod /app/go.mod
COPY ./app/go.sum /app/go.sum

RUN go mod download

COPY --chmod=755 ./asset/entrypoint.sh /entrypoint.sh
COPY ./asset/config.json.template /etc/prism/harness/config.json.template

ENTRYPOINT ["/entrypoint.sh"]

