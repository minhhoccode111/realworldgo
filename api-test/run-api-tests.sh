#!/usr/bin/env bash
set -x

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

APIURL=http://localhost:9999/api/v1
USERNAME=u`date +%s`
EMAIL=$USERNAME@mail.com
PASSWORD='Bruh0!0!'

DELAY_REQUEST=${DELAY_REQUEST:-"500"}

npx newman run $SCRIPTDIR/Conduit.postman_collection.json \
  --delay-request "$DELAY_REQUEST" \
  --global-var "APIURL=$APIURL" \
  --global-var "USERNAME=$USERNAME" \
  --global-var "EMAIL=$EMAIL" \
  --global-var "PASSWORD=$PASSWORD" \
  "$@"
