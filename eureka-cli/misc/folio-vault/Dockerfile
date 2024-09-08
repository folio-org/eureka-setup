ARG VAULT_VERSION=1.17.5

FROM hashicorp/vault:$VAULT_VERSION

RUN apk update && apk add --no-cache curl jq coreutils

COPY config /usr/local/bin/folio/config
COPY scripts /usr/local/bin/folio/scripts

RUN chmod 500 /usr/local/bin/folio/scripts/*
RUN chmod 400 /usr/local/bin/folio/config/*

ENTRYPOINT [ "/usr/local/bin/folio/scripts/start.sh" ]
