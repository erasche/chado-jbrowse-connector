#!/bin/bash

echo "Sleeping on Postgres at db:5432"
until nc -z db 5432; do
    echo "$(date) - waiting for postgres..."
    sleep 2
done

./chado-jb-rest-api \
    --db "postgres://postgres:$POSTGRES_PASSWORD@db/postgres?sslmode=disable" \
    --listenAddr "0.0.0.0:8500" \
    --sitePath "$SITE_PATH" \
    --jbrowse "$JBROWSE"
