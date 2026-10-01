#!/bin/bash

# the initial owner (and its personal access token) can only be created once via the setup endpoint.
# the token is persisted in /opt/netbird.pat (and backed up) and printed for Pulumi: it must never get lost.
# the script is idempotent and also rotates the token, if it expires in less than ROTATE_BEFORE_DAYS days.

PAT_FILE="/opt/netbird.pat"
# the request body of the setup (contains the password): copied by Pulumi, and removed once used
SETUP_BODY_FILE="/opt/netbird/setup.json"
# the lifetime of tokens, and the remaining lifetime at which they are rotated (in days): set by Pulumi
: "${EXPIRY_DAYS:?must be set}"
: "${ROTATE_BEFORE_DAYS:?must be set}"

# extracts a string value from a (compact) JSON document on stdin
json_field() {
    tr -d '\n' | grep -o "\"$1\" *: *\"[^\"]*\"" | head -n 1 | sed 's/^[^:]*: *"\(.*\)"$/\1/'
}

# extracts a value from the yaml token file
pat_field() {
    grep "^$1:" "${PAT_FILE}" | head -n 1 | sed 's/^[^:]*: *//; s/"//g'
}

# writes the token file atomically: user_id email token token_id expires_at
write_pat() {
    (
        umask 077
        {
            echo "---"
            echo "user_id: \"$1\""
            echo "email: \"$2\""
            echo "token: \"$3\""
            echo "token_id: \"$4\""
            echo "expires_at: $5"
        } > "${PAT_FILE}.tmp" && mv "${PAT_FILE}.tmp" "${PAT_FILE}"
    )
    [[ -s "${PAT_FILE}" ]]
}

# wait for netbird to start
until [[ -n "$(docker ps --quiet --filter name=^netbird-server$ --filter status=running)" ]]; do
    echo "Waiting for netbird to start..."
    sleep 5
done
# talk to the container directly (no dns, certificates, or proxy involved)
while true; do
    IP=$(docker inspect netbird-server | grep -o '"IPAddress": "[0-9.][0-9.]*"' | head -n 1 | cut -d '"' -f 4)
    ADDRESS="http://${IP}:80"
    if [[ -n "${IP}" ]]; then
        INSTANCE=$(curl -sf --max-time 10 "${ADDRESS}/api/instance") && [[ -n "${INSTANCE}" ]] && break
    fi
    echo "Waiting for netbird to be reachable..."
    sleep 5
done
CURL="curl -sS --max-time 30"

NOW=$(date +%s)
if [[ -s "${PAT_FILE}" ]]; then
    USER_ID=$(pat_field user_id)
    EMAIL=$(pat_field email)
    TOKEN=$(pat_field token)
    TOKEN_ID=$(pat_field token_id)
    EXPIRES_AT=$(pat_field expires_at)

    if [[ -n "${EXPIRES_AT}" && ${NOW} -lt $((EXPIRES_AT - ROTATE_BEFORE_DAYS * 86400)) ]]; then
        echo "NetBird PAT already exists and is valid. Skipping..."
    else
        echo "Rotating NetBird PAT..."
        RESPONSE=$(${CURL} -X POST "${ADDRESS}/api/users/${USER_ID}/tokens" \
            -H "Authorization: Token ${TOKEN}" -H "Content-Type: application/json" \
            -d "{\"name\":\"pulumi\",\"expires_in\":${EXPIRY_DAYS}}")
        NEW_TOKEN=$(echo "${RESPONSE}" | json_field plain_token)
        NEW_TOKEN_ID=$(echo "${RESPONSE}" | json_field id)

        if [[ -n "${NEW_TOKEN}" ]] && write_pat "${USER_ID}" "${EMAIL}" "${NEW_TOKEN}" "${NEW_TOKEN_ID}" $((NOW + EXPIRY_DAYS * 86400)); then
            # remove the old token, using the new one
            if [[ -n "${TOKEN_ID}" ]]; then
                ${CURL} -X DELETE "${ADDRESS}/api/users/${USER_ID}/tokens/${TOKEN_ID}" \
                    -H "Authorization: Token ${NEW_TOKEN}" > /dev/null || true
            fi
        elif [[ -n "${EXPIRES_AT}" && ${NOW} -lt ${EXPIRES_AT} ]]; then
            echo "Rotating the NetBird PAT failed, but the current one is still valid. Keeping it." >&2
        else
            echo "Rotating the NetBird PAT failed and the current one is expired or unknown." >&2
            echo "Create a personal access token manually and write it to ${PAT_FILE} (yaml: user_id, email, token, token_id, expires_at)." >&2
            exit 1
        fi
    fi
elif echo "${INSTANCE}" | grep -Eq '"setup_required" *: *true'; then
    echo "Setting up NetBird..."
    until [[ -s "${SETUP_BODY_FILE}" ]]; do
        echo "Waiting for the setup request to be uploaded..."
        sleep 5
    done
    RESPONSE=$(${CURL} -X POST "${ADDRESS}/api/setup" -H "Content-Type: application/json" -d "@${SETUP_BODY_FILE}")

    TOKEN=$(echo "${RESPONSE}" | json_field personal_access_token)
    USER_ID=$(echo "${RESPONSE}" | json_field user_id)
    EMAIL=$(echo "${RESPONSE}" | json_field email)
    if [[ -z "${TOKEN}" ]]; then
        echo "NetBird setup did not return a personal access token." >&2
        exit 1
    fi

    # the setup token is the only one: look up its identifier for later rotations (best effort)
    TOKEN_ID=$(${CURL} "${ADDRESS}/api/users/${USER_ID}/tokens" -H "Authorization: Token ${TOKEN}" | json_field id)

    # persist before anything else
    if ! write_pat "${USER_ID}" "${EMAIL}" "${TOKEN}" "${TOKEN_ID}" $((NOW + EXPIRY_DAYS * 86400)); then
        echo "Failed to persist the NetBird PAT." >&2
        exit 1
    fi

    # the request contains the password: it is not needed anymore
    rm -f "${SETUP_BODY_FILE}"
else
    echo "NetBird is already set up, but ${PAT_FILE} does not exist (and could not be restored from backup)." >&2
    echo "Create a personal access token manually and write it to ${PAT_FILE} (yaml: user_id, email, token, token_id, expires_at)." >&2
    exit 1
fi

# output for parsing
echo "--START TOKENS--"
cat "${PAT_FILE}"
echo "--END TOKENS--"
