#!/usr/bin/env bash

docker build -t filesharing_app .
docker run -it --rm -p 8080:8080 filesharing_app
