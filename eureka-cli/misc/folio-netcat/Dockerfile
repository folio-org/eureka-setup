ARG ALPINE_VERSION=3.23

FROM alpine:${ALPINE_VERSION}

RUN apk update && apk add --no-cache netcat-openbsd curl wget bind-tools traceroute && \
    rm -rf /var/cache/apk/* && \
    addgroup -g 1000 netcat && \
    adduser -D -u 1000 -G netcat netcat

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/bin/sh", "-c", "nc -h > /dev/null 2>&1 || exit 1"]

USER 1000:1000

CMD ["/bin/sh", "-c", "sleep infinity"]