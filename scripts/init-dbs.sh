#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE user_db;
    CREATE DATABASE notes_db;
    CREATE DATABASE files_db;

    GRANT ALL PRIVILEGES ON DATABASE user_db TO "$POSTGRES_USER";
    GRANT ALL PRIVILEGES ON DATABASE notes_db TO "$POSTGRES_USER";
    GRANT ALL PRIVILEGES ON DATABASE files_db TO "$POSTGRES_USER";


    -- User Service
    CREATE USER user_service_user WITH PASSWORD '${USER_SERVICE_DB_PASSWORD:-change_me_user}';
    GRANT CONNECT ON DATABASE user_db TO user_service_user;

    -- Notes Service
    CREATE USER notes_service_user WITH PASSWORD '${NOTES_SERVICE_DB_PASSWORD:-change_me_notes}';
    GRANT CONNECT ON DATABASE notes_db TO notes_service_user;

    -- Files Service
    CREATE USER files_service_user WITH PASSWORD '${FILES_SERVICE_DB_PASSWORD:-change_me_files}';
    GRANT CONNECT ON DATABASE files_db TO files_service_user;
EOSQL


psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "user_db" <<-EOSQL
    CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

    REVOKE ALL ON SCHEMA public FROM PUBLIC;

    GRANT USAGE ON SCHEMA public TO user_service_user;
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO user_service_user;
    GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO user_service_user;

    ALTER DEFAULT PRIVILEGES IN SCHEMA public
        GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO user_service_user;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public
        GRANT USAGE, SELECT ON SEQUENCES TO user_service_user;
EOSQL

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "notes_db" <<-EOSQL
    CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

    REVOKE ALL ON SCHEMA public FROM PUBLIC;

    GRANT USAGE ON SCHEMA public TO notes_service_user;
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO notes_service_user;
    GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO notes_service_user;

    ALTER DEFAULT PRIVILEGES IN SCHEMA public
        GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO notes_service_user;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public
        GRANT USAGE, SELECT ON SEQUENCES TO notes_service_user;
EOSQL

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "files_db" <<-EOSQL
    CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

    REVOKE ALL ON SCHEMA public FROM PUBLIC;

    GRANT USAGE ON SCHEMA public TO files_service_user;
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO files_service_user;
    GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO files_service_user;

    ALTER DEFAULT PRIVILEGES IN SCHEMA public
        GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO files_service_user;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public
        GRANT USAGE, SELECT ON SEQUENCES TO files_service_user;
EOSQL

echo "Сервисные пользователи созданы успешно:"
echo "   - user_service_user  -> user_db"
echo "   - notes_service_user -> notes_db"
echo "   - files_service_user -> files_db"