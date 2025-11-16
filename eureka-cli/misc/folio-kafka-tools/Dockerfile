FROM alpine:3.20

RUN apk update && apk add --no-cache wget openjdk17-jre-headless coreutils bash procps-ng && \
    wget -q https://dlcdn.apache.org/kafka/4.1.0/kafka_2.13-4.1.0.tgz && \
    tar -xzf kafka_2.13-4.1.0.tgz && \
    mv kafka_2.13-4.1.0 /opt/kafka && \
    ln -s /opt/kafka/bin/* /usr/local/bin/ && \
    rm kafka_2.13-4.1.0.tgz && \
    rm -rf /var/cache/apk/* && \
    which bash && ls -la /bin/bash && \
    addgroup -g 1000 kafka && \
    adduser -D -u 1000 -G kafka kafka && \
    chown -R kafka:kafka /opt/kafka
    
ENV PATH="/opt/kafka/bin:/usr/bin:/bin:$PATH"

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD test -x /opt/kafka/bin/kafka-topics.sh || exit 1

USER kafka

CMD ["/bin/sleep", "infinity"]