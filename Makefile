.PHONY: all clean protos build-host-agent run-host-agent build-orchestrator run-orchestrator postgres createdb dropdb migrate-up migrate-down
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
	./services/orchestrator/build -config local.yaml
postgres:
	@echo "Running postgres"
	docker run --rm -d --name dbLighthouse -e POSTGRES_USER=dev -e POSTGRES_PASSWORD=dev -p 5432:5432 postgres:13.22-alpine3.22
createdb:
	@echo "Creating database"
	docker exec -it dbLighthouse createdb -U dev lighthouse
dropdb:
	@echo "Dropping database"
	docker exec -it dbLighthouse dropdb -U dev lighthouse
migrate-up:
	@echo "Running migrations up"
	migrate -path ./services/orchestrator/db/migration/ -database "postgres://dev:dev@localhost:5432/lighthouse?sslmode=disable" -verbose up
migrate-down:
	@echo "Running migrations down"
	migrate -path ./services/orchestrator/db/migration/ -database "postgres://dev:dev@localhost:5432/lighthouse?sslmode=disable" -verbose down
sqlc/orchestrator:
	@echo "=> Generating sqlc code for orchestrator..."
	sqlc generate -f services/orchestrator/sqlc.yaml

