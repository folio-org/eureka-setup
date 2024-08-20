ARG KEYCLOAK_VERSION=latest
FROM quay.io/keycloak/keycloak:$KEYCLOAK_VERSION as builder

ENV KC_DB=postgres
ENV KC_CACHE=ispn
ENV KC_HEALTH_ENABLED=true

COPY ./cache-ispn-jdbc.xml /opt/keycloak/conf/cache-ispn-jdbc.xml
ENV KC_CACHE_CONFIG_FILE=cache-ispn-jdbc.xml

COPY ./folio-scripts.jar /opt/keycloak/providers/folio-scripts.jar

RUN /opt/keycloak/bin/kc.sh build --features="scripts,token-exchange,admin-fine-grained-authz"

FROM quay.io/keycloak/keycloak:$KEYCLOAK_VERSION

COPY --from=builder /opt/keycloak/lib/quarkus /opt/keycloak/lib/quarkus

RUN mkdir /opt/keycloak/bin/ebsco
COPY ./ebsco /opt/keycloak/bin/ebsco
COPY ./custom-theme /opt/keycloak/themes/custom-theme
COPY ./folio-scripts.jar /opt/keycloak/providers/folio-scripts.jar

USER root
RUN chmod -R 550 /opt/keycloak/bin/ebsco

USER 1000
ENTRYPOINT [ "/opt/keycloak/bin/ebsco/start.sh", "start", "--optimized" ]
