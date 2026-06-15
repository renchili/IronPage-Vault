FROM golang:1.23-bookworm AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /out/ironpage ./cmd/server

FROM postgres:16-bookworm
COPY --from=builder /out/ironpage /usr/local/bin/ironpage
COPY migrations /opt/ironpage/migrations
COPY public /opt/ironpage/public
COPY scripts/entrypoint.sh /usr/local/bin/ironpage-entrypoint.sh
RUN chmod +x /usr/local/bin/ironpage-entrypoint.sh && mkdir -p /var/lib/ironpage/storage /var/lib/ironpage/backups
ENV POSTGRES_USER=ironpage
ENV POSTGRES_PASSWORD=ironpage
ENV POSTGRES_DB=ironpage
ENV DB_HOST=127.0.0.1
ENV DB_PORT=5432
ENV DB_USER=ironpage
ENV DB_PASSWORD=ironpage
ENV DB_NAME=ironpage
ENV STORAGE_DIR=/var/lib/ironpage/storage
ENV BACKUP_DIR=/var/lib/ironpage/backups
ENV MIGRATIONS_DIR=/opt/ironpage/migrations
ENV HTTP_ADDR=:8080
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/ironpage-entrypoint.sh"]
