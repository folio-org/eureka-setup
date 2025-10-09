FROM ubuntu:22.04

RUN apt-get update && \
    apt-get install -y wget openjdk-17-jre-headless coreutils bash procps && \
    wget -q https://dlcdn.apache.org/kafka/4.1.0/kafka_2.13-4.1.0.tgz && \
    tar -xzf kafka_2.13-4.1.0.tgz && \
    mv kafka_2.13-4.1.0 /opt/kafka && \
    ln -s /opt/kafka/bin/* /usr/local/bin/ && \
    rm kafka_2.13-4.1.0.tgz && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    which bash && ls -la /bin/bash /usr/bin/bash
    
ENV PATH="/opt/kafka/bin:/usr/bin:/bin:$PATH"

CMD ["/usr/bin/sleep", "infinity"]