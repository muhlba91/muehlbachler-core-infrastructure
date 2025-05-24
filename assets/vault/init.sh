#!/bin/bash


ADDRESS="http://127.0.0.1:8200"


# wait for vault to start
until sudo docker inspect vault --format "{{.State.Status}}" | grep "running" > /dev/null; do
    echo "Waiting for vault to start..."
    sleep 5
done
while [[ ! $(curl -s ${ADDRESS}) ]]; do
    echo "Waiting for vault to be reachable..."
    sleep 5
done


# initialize vault
if [[ -f /opt/vault-init.txt ]]; then
    echo "Vault already initialized. Skipping..."
else
    echo "Initializing vault..."
    sudo docker exec vault /bin/sh -c "vault operator init --address '${ADDRESS}'" | sudo tee /opt/vault-init.txt > /dev/null
fi

# check if vault was already initialized or not
if [[ $? -eq 0 ]]; then
    # parse token and recovery keys
    echo "Vault initialized successfully. Parsing keys..."
    VAULT_TOKEN=$(sudo grep 'Initial Root Token:' /opt/vault-init.txt | awk '{print $4}')
    VAULT_RECOVERY_KEY_1=$(sudo grep 'Recovery Key 1:' /opt/vault-init.txt | awk '{print $4}')
    VAULT_RECOVERY_KEY_2=$(sudo grep 'Recovery Key 2:' /opt/vault-init.txt | awk '{print $4}')
    VAULT_RECOVERY_KEY_3=$(sudo grep 'Recovery Key 3:' /opt/vault-init.txt | awk '{print $4}')
    VAULT_RECOVERY_KEY_4=$(sudo grep 'Recovery Key 4:' /opt/vault-init.txt | awk '{print $4}')
    VAULT_RECOVERY_KEY_5=$(sudo grep 'Recovery Key 5:' /opt/vault-init.txt | awk '{print $4}')

    # output for parsing
    echo "--START TOKENS--"
    echo "---"
    echo "root_token: $VAULT_TOKEN"
    echo "recovery_keys:"
    echo "  - ${VAULT_RECOVERY_KEY_1}"
    echo "  - ${VAULT_RECOVERY_KEY_2}"
    echo "  - ${VAULT_RECOVERY_KEY_3}"
    echo "  - ${VAULT_RECOVERY_KEY_4}"
    echo "  - ${VAULT_RECOVERY_KEY_5}"
    echo "--END TOKENS--"
else
    echo "Vault already initialized. Skipping..."
fi
