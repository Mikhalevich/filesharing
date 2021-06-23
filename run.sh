#!/usr/bin/env bash

SCRIPT_DIR=$(dirname "$0")
pushd $SCRIPT_DIR

if [ ! -f "cert_auth/private_key.pem" ] || [ ! -f "cert_auth/public_key.pem" ]
then
    pushd $SCRIPT_DIR/cert_auth
    ./generate.sh
    popd
fi

docker-compose up --build

popd