FROM alpine:3.20

RUN apk update && apk add --no-cache \
    netcat-openbsd=1.226-r0 \
    curl=8.14.1-r2 \
    wget=1.24.5-r0 \
    bind-tools=9.18.39-r0 \
    traceroute=2.1.5-r0 && \
    rm -rf /var/cache/apk/*

CMD ["/bin/sh", "-c", "sleep infinity"]