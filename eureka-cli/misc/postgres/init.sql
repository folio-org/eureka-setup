ALTER SYSTEM SET max_connections = 500;

-- System
create database keycloak;
create user keycloak_rw with password 'supersecret';
grant connect on database keycloak to keycloak_rw;
grant all privileges on database keycloak to keycloak_rw;

create database kong;
create user kong_rw with password 'supersecret';
grant connect on database kong to kong_rw;
grant all privileges on database kong to kong_rw;

-- Management
create database mgr_applications;
create user mgr_applications_rw with password 'supersecret';
grant connect on database mgr_applications to mgr_applications_rw;
grant all privileges on database mgr_applications to mgr_applications_rw;

create database mgr_tenant;
create user mgr_tenant_rw with password 'supersecret';
grant connect on database mgr_tenant to mgr_tenant_rw;
grant all privileges on database mgr_tenant to mgr_tenant_rw;

create database mgr_tenant_entitlements;
create user mgr_tenant_entitlements_rw with password 'supersecret';
grant connect on database mgr_tenant_entitlements to mgr_tenant_entitlements_rw;
grant all privileges on database mgr_tenant_entitlements to mgr_tenant_entitlements_rw;

-- Core and other modules
create database folio;
create user folio_rw with password 'supersecret';
alter user folio_rw with superuser;