BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id varchar(64) primary key not null,
    created_at timestamp without time zone default now() not null,

    username varchar(64) not null unique,
    avatar varchar(512),
    sex varchar(32) not null,
    email varchar(64) not null
);

CREATE INDEX IF NOT EXISTS users_created_at_idx ON users USING btree(created_at);

COMMIT;