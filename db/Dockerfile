# Use the official lightweight Alpine image
FROM alpine:3.20

RUN apk add --no-cache postgresql postgresql-client

ENV DATA_DIR=/var/lib/postgresql/data \
    POSTGRES_DB=mydb \
    POSTGRES_USER=myuser \
    POSTGRES_PASSWORD=mypassword

RUN mkdir -p "$DATA_DIR" /run/postgresql && \
    chown -R postgres:postgres "$DATA_DIR" /run/postgresql

EXPOSE 5432

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

USER postgres

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
