create table if not exists counters (id serial primary key, name varchar(255) unique, value double precision);