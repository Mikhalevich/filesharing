#!/usr/bin/env bash

protoc -I proto/ --go_out=paths=source_relative:./proto --micro_out=paths=source_relative:./proto types/types.proto
protoc -I proto/ --go_out=paths=source_relative:./proto --micro_out=paths=source_relative:./proto auth/auth.proto
protoc -I proto/ --go_out=paths=source_relative:./proto --micro_out=paths=source_relative:./proto file/file.proto
