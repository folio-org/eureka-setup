ALTER SYSTEM SET max_connections = 500;

create database okapi;
create user okapi_rw with password 'superAdmin';
alter user okapi_rw with superuser;

create database app_manager;
create user app_manager_rw with password 'superAdmin';
grant connect on database app_manager to app_manager_rw;
grant all privileges on database app_manager to app_manager_rw;

create database tenant_manager;
create user tenant_manager_rw with password 'superAdmin';
grant connect on database tenant_manager to tenant_manager_rw;
grant all privileges on database tenant_manager to tenant_manager_rw;

create database tenant_entitlement;
create user tenant_entitlement_rw with password 'superAdmin';
grant connect on database tenant_entitlement to tenant_entitlement_rw;
grant all privileges on database tenant_entitlement to tenant_entitlement_rw;

create database keycloak;
create user keycloak_rw with password 'superAdmin';
grant connect on database keycloak to keycloak_rw;
grant all privileges on database keycloak to keycloak_rw;

create database kong;
create user kong_rw with password 'superAdmin';
grant connect on database kong to kong_rw;
grant all privileges on database kong to kong_rw;
