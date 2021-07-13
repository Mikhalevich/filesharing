#!/usr/bin/env bash

protoc -I proto/auth --go_out=proto/auth --micro_out=proto/auth proto/auth/auth.proto
protoc -I proto/file --go_out=proto/file --micro_out=proto/file proto/file/file.proto
