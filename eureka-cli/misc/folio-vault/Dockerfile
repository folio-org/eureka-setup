ARG VAULT_VERSION=1.21

FROM hashicorp/vault:$VAULT_VERSION

RUN apk update && apk add --no-cache curl jq coreutils && \
    rm -rf /var/cache/apk/* && \
    addgroup -S folio && \
    adduser -S -G folio folio

COPY config /usr/local/bin/folio/config
COPY scripts /usr/local/bin/folio/scripts

RUN chmod 500 /usr/local/bin/folio/scripts/* && \
    chmod 400 /usr/local/bin/folio/config/* && \
    chown -R folio:folio /usr/local/bin/folio

USER folio

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8200/v1/sys/health || exit 1

ENTRYPOINT [ "/usr/local/bin/folio/scripts/start.sh" ]

