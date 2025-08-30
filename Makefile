.PHONY: all clean protos build-host-agent run-host-agent build-orchestrator run-orchestrator postgres createdb dropdb migrate-up migrate-down
protos:
	@echo "Building proto for host-agent"
	protoc -I proto \
       --go_out=services/common/genproto/host-agents/ \
       --go_opt=paths=source_relative \
       --go-grpc_out=services/common/genproto/host-agents/ \
       --go-grpc_opt=paths=source_relative \
       proto/host-agent.proto
	@echo "Building proto for registry-monitor"
	protoc -I proto \
			 --go_out=services/common/genproto/registry-monitor/ \
			 --go_opt=paths=source_relative \
			 --go-grpc_out=services/common/genproto/registry-monitor/ \
			 --go-grpc_opt=paths=source_relative \
			 proto/registry-monitor.proto



run-host-agent:
	@echo "Running host-agent"
	go build -o ./services/host-agent/build ./services/host-agent/cmd/agent/main.go 
	./services/host-agent/build

run-orchestrator:
	@echo "Running orchestrator"
	go build -o ./services/orchestrator/build ./services/orchestrator/cmd/orchestrator/main.go
	./services/orchestrator/build -config local.yaml
run-registry-monitor:
	@echo "Running registry-monitor"
	go build -o ./services/registry-monitor/build ./services/registry-monitor/cmd/monitor/main.go
	./services/registry-monitor/build
run-tui:
	@echo "Running TUI"
	go build -o ./services/tui/build ./services/tui/cmd/tui/main.go
	./services/tui/build
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

make-db:
	@echo "Setting up the database"
	make postgres
	make createdb
	make migrate-up

