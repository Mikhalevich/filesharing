#!/usr/bin/env bash

SCRIPT_DIR=$(dirname "$0")
pushd $SCRIPT_DIR

if [ ! -f "filesharing-auth-service/token/cert/private_key.pem" ] || [ ! -f "filesharing-auth-service/token/cert/public_key.pem" ]
then
    pushd $SCRIPT_DIR/filesharing-auth-service/token/cert
    ./generate.sh
    popd
fi

docker-compose up --build --remove-orphans

popd