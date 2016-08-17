#!/bin/bash

echo "Sleeping on Postgres at db:5432"
until nc -z db 5432; do
    echo "$(date) - waiting for postgres..."
    sleep 2
done

export CHADOJB_DBSTRING="postgres://postgres:$POSTGRES_PASSWORD@db/postgres?sslmode=disable"
export CHADOJB_LISTENADDR="0.0.0.0:8500"
export CHADOJB_SITEPATH="$SITE_PATH"
export CHADOJB_JBROWSE="$JBROWSE"

./chado-jb-rest-api
