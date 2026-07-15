FROM golang:1.25-bookworm AS builder
WORKDIR /src

# This builder resolves modules and generates Swagger internally. The source
# checkout does not need go.sum or docs/swagger artifacts.
COPY go.mod ./
RUN go mod download
COPY . .
RUN sh scripts/build_server_in_container.sh

FROM postgres:16-bookworm
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        poppler-utils \
        python3 \
        python3-pip \
        tar \
    && python3 -m pip install --break-system-packages --no-cache-dir pypdf reportlab pillow \
    && rm -rf /var/lib/apt/lists/*
COPY --from=builder /out/ironpage /usr/local/bin/ironpage
COPY migrations /opt/ironpage/migrations
COPY public /opt/ironpage/public
COPY scripts/entrypoint.sh /usr/local/bin/ironpage-entrypoint.sh
RUN chmod +x /usr/local/bin/ironpage-entrypoint.sh \
    && mkdir -p /var/lib/ironpage/storage /var/lib/ironpage/backups

# Runtime database identity, credentials, signing material, encryption material,
# and acceptance identities must be supplied by the deployment environment.
ENV STORAGE_DIR=/var/lib/ironpage/storage
ENV BACKUP_DIR=/var/lib/ironpage/backups
ENV MIGRATIONS_DIR=/opt/ironpage/migrations
ENV PUBLIC_DIR=/opt/ironpage/public
ENV HTTP_ADDR=:8080

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/ironpage-entrypoint.sh"]
