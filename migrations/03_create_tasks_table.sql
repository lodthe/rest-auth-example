BEGIN;

CREATE TABLE IF NOT EXISTS tasks (
      id varchar(64) primary key not null,
      created_at timestamp without time zone default now() not null,

      user_id varchar(64) not null references users(id),

      status integer default 0 not null,
      result jsonb
);

CREATE INDEX IF NOT EXISTS tasks_created_at_idx ON tasks USING btree(created_at);
CREATE INDEX IF NOT EXISTS tasks_status_idx ON tasks USING btree(status);

COMMIT;