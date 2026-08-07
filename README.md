# Dependencies

## Goose

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## SQLC

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

## Make

### MacOS

#### Using CLI Tools

```bash
xcode-select --install
```

#### Using Homebrew

```bash
brew install make
```

### Linux

#### Ubuntu/Debian:

```bash
sudo apt install make
```

#### Fedora/RHEL/CentOS

```bash
sudo dnf install make
```

#### Arch

```bash
sudo pacman -S make
```

### Windows

#### Powershell/CMD

Run as administrator.

```bash
winget install GnuWin32.Make
```

#### Chocolatey

```bash
choco install make
```

# Create Database

1. Navigate to `internal/db`

```bash
cd internal/db
```

2. Create a `.env` file, use the template provided and fill in the actual values you want your database to use.

3. Run the docker compose command

```bash
docker compose up -d
```

# Apply Migrations

1. Create a `.env` file at the project root, use the template provided and replace the mockup values with your actual credentials.

2. Navigate to `internal/db/migrations`

```bash
cd internal/db/migrations
```

3. Run the desired migration command

To update your database to the latest unsynced migration:

```bash
make migrate-up
```

To revert your database one migration at a time:

```bash
make migrate-down
```

To check the current migration status:

```bash
make migrate-status
```

# Generate Go code from queries

SQL Queries must be placed in `db/queries`. Once you are done writing custom queries, you can generate Go code from them using SQLC.

Run this command at the project root.

```bash
make sqlc
```

This will place the generated code in `internal/store/sqlc`. Do not modify these files manually.
