create database konga;
create user konga_rw with password 'superAdmin';
grant connect on database konga to konga_rw;
grant all privileges on database konga to konga_rw;
