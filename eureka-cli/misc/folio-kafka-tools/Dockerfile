ARG ALPINE_VERSION=3.23

FROM alpine:${ALPINE_VERSION}

ARG KAFKA_VERSION=4.1.1
ARG SCALA_VERSION=2.13

RUN apk update && apk add --no-cache wget openjdk17-jre-headless coreutils bash procps-ng && \
    wget -q https://dlcdn.apache.org/kafka/${KAFKA_VERSION}/kafka_${SCALA_VERSION}-${KAFKA_VERSION}.tgz && \
    tar -xzf kafka_${SCALA_VERSION}-${KAFKA_VERSION}.tgz && \
    mv kafka_${SCALA_VERSION}-${KAFKA_VERSION} /opt/kafka && \
    ln -s /opt/kafka/bin/* /usr/local/bin/ && \
    rm kafka_${SCALA_VERSION}-${KAFKA_VERSION}.tgz && \
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