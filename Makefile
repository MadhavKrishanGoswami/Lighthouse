protos:
	@echo "Building proto for host-agent"
	protoc -I proto \
       --go_out=services/common/genproto/host-agents/ \
       --go_opt=paths=source_relative \
       --go-grpc_out=services/common/genproto/host-agents/ \
       --go-grpc_opt=paths=source_relative \
       proto/host-agent.proto


build-host-agent:
	@echo "Building host-agent binary"
	go build -o ./services/host-agent/build ./services/host-agent/cmd/agent/main.go 

run-host-agent:
	@echo "Running host-agent"
	./services/host-agent/build

build-orchestrator:
	@echo "Building orchestrator binary"
	go build -o ./services/orchestrator/build ./services/orchestrator/cmd/orchestrator/main.go
run-orchestrator:
	@echo "Running orchestrator"
	./services/orchestrator/build -config ./services/orchestrator/internal/config/local.yaml
