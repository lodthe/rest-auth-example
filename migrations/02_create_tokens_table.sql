BEGIN;

CREATE TABLE IF NOT EXISTS tokens (
    id varchar(64) primary key not null,

    type varchar(32) not null,
    parent_id varchar(64),

    issued_at timestamp without time zone default now() not null,
    expires_at timestamp without time zone,

    user_id varchar(64) not null references users(id)
);

CREATE INDEX IF NOT EXISTS tokens_user_id_idx ON tokens USING btree(user_id);
CREATE INDEX IF NOT EXISTS tokens_parent_id_idx ON tokens USING btree(parent_id);

COMMIT;