#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER hedera_db_validation with PASSWORD 'hedera_db_validation_pass';
EOSQL