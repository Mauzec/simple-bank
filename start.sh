#!/bin/sh
set -e

echo "run db migrate"
/usr/bin/migrate -path /app/migrate -database "$DB_SOURCE" -verbose up

echo "start the app"
exec "$@"