ARG VAULT_VERSION=1.19

FROM hashicorp/vault:$VAULT_VERSION

RUN apk update && apk add --no-cache \
    curl=8.14.1-r2 \
    jq=1.7.1-r0 \
    coreutils=9.5-r2

# Create a non-root user and group
RUN addgroup -S folio && adduser -S -G folio folio

# Create and set permissions on directories
COPY config /usr/local/bin/folio/config
COPY scripts /usr/local/bin/folio/scripts

RUN chmod 500 /usr/local/bin/folio/scripts/* && \
    chmod 400 /usr/local/bin/folio/config/* && \
    chown -R folio:folio /usr/local/bin/folio

# Switch to non-root user
USER folio

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8200/v1/sys/health || exit 1

ENTRYPOINT [ "/usr/local/bin/folio/scripts/start.sh" ]
