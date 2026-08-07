# Dependencies

## Goose

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

# Create Database

1. Navigate to internal/db

```bash
cd internal/db
```

2. Create a .env file and fill in the missing values. (Use .env.template and remember to update the connection string, the one provided is just a template)

Example:

```.env
POSTGRES_DB="dev_db"
POSTGRES_USER="admin"
POSTGRES_PASSWORD="ultrasecurepassword"
POSTGRESS_CONNECTION_STRING="postgres://admin:ultrasecurepassword@localhost:5342/dev_db"
```

3. Run the docker compose command

```bash
docker compose up -d
```

# Apply Migrations

1. Navigate to internal/db/migrations

```bash
cd internal/db/migrations
```

2. Run the up migration command

```bash
goose postgres {CONN_STRING} up
```

Replace `{CONN_STRING}` with the actual connection string of your database.
