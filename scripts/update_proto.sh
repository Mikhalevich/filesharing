#!/usr/bin/env bash

protoc -I pkg/proto/ --go_out=paths=source_relative:./pkg/proto --micro_out=paths=source_relative:./pkg/proto types/types.proto
protoc -I pkg/proto/ --go_out=paths=source_relative:./pkg/proto --micro_out=paths=source_relative:./pkg/proto auth/auth.proto
protoc -I pkg/proto/ --go_out=paths=source_relative:./pkg/proto --micro_out=paths=source_relative:./pkg/proto file/file.proto
protoc -I pkg/proto/ --go_out=paths=source_relative:./pkg/proto --micro_out=paths=source_relative:./pkg/proto event/event.proto
protoc -I pkg/proto/ --go_out=paths=source_relative:./pkg/proto --micro_out=paths=source_relative:./pkg/proto history/history.proto
