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

# Endpoints

## Resources

### Users

#### `POST /api/users`

Creates a user. It handles password validation with the following rules:

- Be 8 characters long.
- Contain at lest one uppercase letter.
- Contain at leat one lowercase letter.
- Contain at least one digit.
- Contain at least one of the following special characters: `@$!%\*?&`.

If the password is valid then it is hashed before being inserted into the database.

Example request body:

```json
{
  "email": "xcxsar@github.co",
  "password": "passW0rd123!"
}
```

Example response:

```json
{
  "id": "11623d71-9a47-4fec-87a4-023f607af30d",
  "created_at": "2026-08-07T22:56:16.001797Z",
  "updated_at": "2026-08-07T22:56:16.001797Z",
  "email": "xcxsar@github.co"
}
```

#### `GET /api/users/{userID}`

Gets a user by the provided ID. No body is required.

Example URL:

`http://localhost:8080/api/users/11623d71-9a47-4fec-87a4-023f607af30d`

Example response:

```json
{
  "id": "11623d71-9a47-4fec-87a4-023f607af30d",
  "created_at": "2026-08-07T22:56:16.001797Z",
  "updated_at": "2026-08-07T22:56:16.001797Z",
  "email": "xcxsar@gmail.com"
}
```

#### `PATCH /api/user/email`

Updates a user's email. Bearer token authorization header is required.

Example request body:

```json
{
  "email": "xcxs4r@gmail.com"
}
```

Example response:

```json
{
  "id": "11623d71-9a47-4fec-87a4-023f607af30d",
  "created_at": "2026-08-07T22:56:16.001797Z",
  "updated_at": "2026-08-14T00:13:20.09183Z",
  "email": "xcxs4r@gmail.com"
}
```

#### `PATCH /api/user/password`

Updates a user's password. Bearer token authorization header is required.

Example request body:

```json
{
  "password": "passW000rd123!"
}
```

Example response:

```json
{
  "id": "11623d71-9a47-4fec-87a4-023f607af30d",
  "created_at": "2026-08-07T22:56:16.001797Z",
  "updated_at": "2026-08-14T00:18:15.941573Z",
  "email": "xcxs4r@gmail.com"
}
```
