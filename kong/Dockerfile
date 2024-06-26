ARG KONG_VERSION=3.4.1-ubuntu
FROM kong:$KONG_VERSION

ENV KONG_PROXY_ACCESS_LOG="/dev/stdout txns"
# This effectively means that both proxy access and admin access logs will both be sent to stdout
# and log entries from both logs will be comingled.  While not ideal, it's sufficient for now.  We
# should revisit this later to see if separating them is worthwhile, and what our options are.
ENV KONG_ADMIN_ACCESS_LOG="/dev/stdout txns"
ENV KONG_NGINX_HTTP_LOG_FORMAT="txns '\$http_x_forwarded_for - \$remote_addr - \$remote_user [\$time_local] \"\$request\" \$status \$body_bytes_sent rt=\$request_time uct=\"\$upstream_connect_time\" uht=\"\$upstream_header_time\" urt=\"\$upstream_response_time\" \"\$http_user_agent\" \"\$http_x_okapi_tenant\"'"
ENV KONG_ROUTER_FLAVOR=expressions

USER root

ARG TARGETARCH
ARG DECK_VERSION=1.27.1
ARG DECK_DIRECTORY="/usr/local/bin/deck"
ARG DECK_ARTIFACT_NAME="deck_${DECK_VERSION}_linux_${TARGETARCH}.tar.gz"
ARG DECK_DOWNLOAD_URL="https://github.com/kong/deck/releases/download/v${DECK_VERSION}/${DECK_ARTIFACT_NAME}"

RUN apt-get update \
    && apt-get -y upgrade \
    && apt-get install -y curl zip \
    && apt-get clean

RUN mkdir "$DECK_DIRECTORY" \
    && curl -sL "$DECK_DOWNLOAD_URL" -o "$DECK_DIRECTORY/$DECK_ARTIFACT_NAME" \
    && tar -xf "$DECK_DIRECTORY/$DECK_ARTIFACT_NAME" -C "$DECK_DIRECTORY" \
    && rm -f "$DECK_DIRECTORY/$DECK_ARTIFACT_NAME"
ENV PATH = "$PATH:$DECK_DIRECTORY"
COPY config /opt/kong/config
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /usr/local/bin/deck /entrypoint.sh
RUN chown -R kong /opt/kong

# Insall auth-headers-manager plugin
ARG AUTH_HEADERS_MANAGER_PLUGIN_NAME=auth-headers-manager
ARG AUTH_HEADERS_MANAGER_PLUGIN_VERSION="1.0-1"
COPY /$AUTH_HEADERS_MANAGER_PLUGIN_NAME /usr/local/plugins/$AUTH_HEADERS_MANAGER_PLUGIN_NAME

RUN cd /usr/local/plugins/$AUTH_HEADERS_MANAGER_PLUGIN_NAME \
    && luarocks make "kong-plugin-$AUTH_HEADERS_MANAGER_PLUGIN_NAME-$AUTH_HEADERS_MANAGER_PLUGIN_VERSION.rockspec" \
    && cd / \
    && rm -rf /user/local/plugins/$AUTH_HEADERS_MANAGER_PLUGIN_NAME

USER kong

ENTRYPOINT ["/entrypoint.sh"]
CMD ["kong", "docker-start"]
