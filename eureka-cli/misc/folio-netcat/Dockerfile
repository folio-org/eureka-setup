FROM alpine:3.20

RUN apk update && apk add --no-cache netcat-openbsd curl wget bind-tools traceroute && \
    rm -rf /var/cache/apk/* && \
    addgroup -g 1000 netcat && \
    adduser -D -u 1000 -G netcat netcat

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD nc -h > /dev/null 2>&1 || exit 1

USER netcat

CMD ["/bin/sh", "-c", "sleep infinity"]