.PHONY = clean grpc

all:
	cd cmd/cmmonitor && go build

grpc: internal/api/cmts.pb.go clients/php/src/Cmts

internal/api/cmts.pb.go: api/cmts.proto
	protoc --go_out=plugins=grpc:internal/api api/cmts.proto

clients/php/src/Cmts:
	protoc --proto_path=api --php_out=clients/php/src --grpc_out=clients/php/src --plugin=protoc-gen-grpc=/usr/bin/grpc_php_plugin ./api/cmts.proto

#clean:
#	rm rpc/*.pb.go
