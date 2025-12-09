#!/bin/bash
docker exec minio mc alias set local http://localhost:9000 minioadmin minioadmin
docker exec minio mc mb local/notes-app --ignore-existing
docker cp notes_service/icons minio:/tmp/icons
docker exec minio mc mirror /tmp/icons local/notes-app/icons
docker exec minio rm -rf /tmp/icons
