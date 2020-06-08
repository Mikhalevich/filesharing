#!/usr/bin/env bash

if [ ! -f "private_key.pem" ] || [ "$1" = "-f" ]
then
    openssl genpkey -algorithm RSA -out private_key.pem -pkeyopt rsa_keygen_bits:2048
    openssl rsa -pubout -in private_key.pem -out public_key.pem
fi


