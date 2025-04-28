create database keycloak;
create user keycloak_rw with password 'supersecret';
grant connect on database keycloak to keycloak_rw;
grant all privileges on database keycloak to keycloak_rw;

create database kong;
create user kong_rw with password 'supersecret';
grant connect on database kong to kong_rw;
grant all privileges on database kong to kong_rw;

create database folio;
create user folio_rw with password 'supersecret';
alter user folio_rw with superuser;

-- Change schema owner
-- https://www.postgresql.org/docs/15/release-15.html

\connect keycloak;
alter schema public owner to keycloak_rw;

\connect kong;
alter schema public owner to kong_rw;

\connect folio;
alter schema public owner to folio_rw;