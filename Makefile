protos:
	@echo "Building proto for host-agent"
	protoc -I proto \
       --go_out=services/common/genproto/host-agents/ \
       --go_opt=paths=source_relative \
       --go-grpc_out=services/common/genproto/host-agents/ \
       --go-grpc_opt=paths=source_relative \
       proto/host-agent.proto

