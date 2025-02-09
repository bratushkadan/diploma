#!/bin/sh

set -e

FILE=.ydb-access-token

if [ ! -f "${FILE}" ]; then
    LOCKBOX_SEC_ID="$(../terraform/tf output -json | jq -cMr .app_sa.value.auth_key_lockbox_secret_id)"
    TMP_FILE="$(mktemp)"
    PRIVATE_KEY=$(yc lockbox payload get --format json "${LOCKBOX_SEC_ID}" \
       | jq -cMr '.entries | .[] | select(.key == "auth_key").text_value') 
    yc iam key get --format json --full \
        "$(../terraform/tf output -json | jq -cMr .app_sa.value.key_id)" \
        | jq -Mr ".  + {\"private_key\": \"${PRIVATE_KEY}\"}" > "${TMP_FILE}"
    jq -Mr ".  + {\"private_key\": \"${PRIVATE_KEY}\"}" "${TMP_FILE}" > "${FILE}"
    rm "${TMP_FILE}"
fi

echo "${FILE}"

