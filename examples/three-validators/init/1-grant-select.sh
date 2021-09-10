#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON tables TO hedera_db_validation;
    GRANT SELECT ON ALL TABLES IN SCHEMA public TO hedera_db_validation;
EOSQL