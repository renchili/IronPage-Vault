FROM golang:1.25-bookworm AS builder
WORKDIR /src

# This builder resolves modules and generates Swagger internally. The source
# checkout does not need go.sum or docs/swagger artifacts.
COPY go.mod ./
RUN go mod download
COPY . .
RUN sh scripts/build_server_in_container.sh

FROM postgres:16-bookworm
ARG IRONPAGE_APP_ROOT
ARG IRONPAGE_HTTP_PORT
RUN test -n "$IRONPAGE_APP_ROOT" \
    && test -n "$IRONPAGE_HTTP_PORT" \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
        poppler-utils \
        python3 \
        python3-pip \
        tar \
    && python3 -m pip install --break-system-packages --no-cache-dir pypdf reportlab pillow \
    && rm -rf /var/lib/apt/lists/*
ENV IRONPAGE_APP_ROOT=${IRONPAGE_APP_ROOT}
ENV IRONPAGE_HTTP_PORT=${IRONPAGE_HTTP_PORT}
COPY --from=builder /out/ironpage /usr/local/bin/ironpage
COPY migrations ${IRONPAGE_APP_ROOT}/migrations
COPY public ${IRONPAGE_APP_ROOT}/public
COPY scripts/entrypoint.sh /usr/local/bin/ironpage-entrypoint.sh
RUN chmod +x /usr/local/bin/ironpage-entrypoint.sh

# Runtime database identity, credentials, ports, filesystem locations, signing
# material, encryption material, and acceptance identities are supplied by the
# generated deployment environment. The image contains no fixed local runtime
# configuration.
EXPOSE ${IRONPAGE_HTTP_PORT}
ENTRYPOINT ["/usr/local/bin/ironpage-entrypoint.sh"]
