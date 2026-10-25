# SQLite to PostgreSQL

The PostgreSQL Compose service is a new persistent volume. It does not copy or delete the existing SQLite database automatically.

1. Stop the Go server so no writes happen during the snapshot: `docker compose -f deploy/docker-compose/docker-compose.yaml stop server`.
2. Start only PostgreSQL: `docker compose -f deploy/docker-compose/docker-compose.yaml up -d postgres`.
3. Install `pgloader` and run `./script/migrate_sqlite_to_pg.sh`.
4. Verify row counts and credentials before starting the full stack.

The script creates an online SQLite backup under the ignored root `tmp/` directory, then imports that snapshot. Keep the original SQLite file until the PostgreSQL data has been checked.

The default host connection is `postgresql://xianyu:xianyu-local-only@127.0.0.1:15432/xianyu`. Override the second script argument when `PGSQL_PASSWORD` is customized.
