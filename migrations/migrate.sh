#!/bin/sh
set -e

DBSTRING="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"
VERSION="${MIGRATION_VERSION:-latest}"

echo "applying migrations up to version: ${VERSION}"

DIR="$(cd "$(dirname "$0")" && pwd)"
MIGRATIONS_DIR="${DIR}/sql"

if [ "$VERSION" = "latest" ]; then
    goose -dir "$MIGRATIONS_DIR" postgres "$DBSTRING" up
else
    goose -dir "$MIGRATIONS_DIR" postgres "$DBSTRING" up-to "$VERSION"
fi