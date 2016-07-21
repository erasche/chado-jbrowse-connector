#!/bin/sh

./chado-jb-rest-api \
    --db "postgres://postgres:$POSTGRES_PASSWORD@db/postgres?sslmode=disable" \
    --listenAddr "0.0.0.0:8500" \
    --sitePath "$SITE_PATH"
