#!/bin/sh
set -e

if [ -z "$(ls -A "$DATA_DIR")" ]; then
    echo "Initializing database..."
    initdb -D "$DATA_DIR"

    pg_ctl -D "$DATA_DIR" -o "-k /tmp" -w start

    echo "Configuring roles and databases..."
    psql --username=postgres --host=/tmp -c "CREATE USER \"$POSTGRES_USER\" WITH PASSWORD '$POSTGRES_PASSWORD';"
    psql --username=postgres --host=/tmp -c "CREATE DATABASE \"$POSTGRES_DB\" OWNER \"$POSTGRES_USER\";"
    psql --username=postgres --host=/tmp -c "GRANT ALL PRIVILEGES ON DATABASE \"$POSTGRES_DB\" TO \"$POSTGRES_USER\";"

    echo "host all all 0.0.0.0/0 md5" >> "$DATA_DIR/pg_hba.conf"
    echo "listen_addresses='*'" >> "$DATA_DIR/postgresql.conf"
    echo "unix_socket_directories='/tmp'" >> "$DATA_DIR/postgresql.conf"

    pg_ctl -D "$DATA_DIR" -o "-k /tmp" -m fast stop
fi

exec postgres -D "$DATA_DIR" -c unix_socket_directories='/tmp'
