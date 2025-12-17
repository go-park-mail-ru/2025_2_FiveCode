#!/bin/sh
set -e

echo "Waiting for MinIO to be ready..."
until mc alias set myminio http://$MINIO_HOST:$MINIO_PORT $MINIO_ACCESS_KEY $MINIO_SECRET_KEY; do
  echo "MinIO not ready yet... sleeping 2s"
  sleep 2
done

echo "MinIO is ready."

mc mb myminio/notes-app --ignore-existing

mc anonymous set download myminio/notes-app

echo "Uploading icons..."
mc mirror --overwrite /icons myminio/notes-app/icons

echo "Icons uploaded successfully!"

echo "Uploading headers..."
mc mirror --overwrite /headers myminio/notes-app/headers

echo "Headers uploaded successfully!"