#!/bin/sh
set -e

if [ -n "$FCM_CREDENTIALS_JSON_B64" ]; then
  mkdir -p /app/credentials
  echo "$FCM_CREDENTIALS_JSON_B64" | base64 -d > /app/credentials/firebase.json
  export FCM_CREDENTIALS_PATH=/app/credentials/firebase.json
fi

exec ./notification-service