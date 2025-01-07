FROM prism/prism-go

RUN apt update \
	&& apt -y install \
		libgmp-dev \
		build-essential \
		flex \
		bison \
		vim \
		jq

WORKDIR /opt/prism-harness

COPY ./app/go.mod /opt/prism-harness/go.mod
COPY ./app/go.sum /opt/prism-harness/go.sum

RUN go mod download

COPY --chmod=755 ./asset/entrypoint.sh /entrypoint.sh
COPY ./asset/config.json.template /etc/prism/harness/config.json.template

COPY ./app /opt/prism-harness
RUN go build -o /usr/local/bin/harness ./cmd/harness
RUN go build -o /usr/local/bin/aide ./cmd/aide

ENV PRISM_HARNESS_DIR=/opt/prism-harness
ENV PRISM_HARNESS_DATA_DIR=/var/lib/prism-harness

ENTRYPOINT ["/entrypoint.sh"]

