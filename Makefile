NAME = github.com/pavelborisofff/go-metrics

.PHONY: go-init
go-init:
	go mod init $(NAME)

.PHONY: go-run-server
go-run-server:
	go run cmd/server/main.go

.PHONY: go-run-server-race
go-run-server-race:
	go run -race cmd/server/main.go

.PHONY: go-run-agent
go-run-agent:
	go run cmd/agent/main.go

.PHONY: go-run-agent-race
go-run-agent-race:
	go run -race cmd/agent/main.go

.PHONY: go-run-tests
go-run-tests:
	go test ./...

.PHONY: git-checkout
git-checkout:
	git checkout -b $(BRANCH)

.PHONY: go-run-autotests-2
go-run-autotests-2:
	go build -o cmd/server/server cmd/server/main.go
	go build -o cmd/agent/agent cmd/agent/main.go
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration2[AB]*$$ -source-path=. -agent-binary-path=cmd/agent/agent

.PHONY: go-run-autotests-3
go-run-autotests-3:
	go build -o cmd/server/server cmd/server/main.go
	go build -o cmd/agent/agent cmd/agent/main.go
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration3[AB]*$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server

.PHONY: go-run-autotests-10
go-run-autotests-10:
	go build -o cmd/server/server cmd/server/main.go
	go build -o cmd/agent/agent cmd/agent/main.go
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration10[AB]$$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=12345 -source-path=. -file-storage-path=TEMP_FILE -database-dsn='postgres://postgres:password@localhost:15432/praktikum?sslmode=disable'

.PHONY: go-run-autotests
go-run-autotests:
	go build -o cmd/server/server cmd/server/main.go
	go build -o cmd/agent/agent cmd/agent/main.go
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration${TEST_NUM}$$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=12345 -source-path=. -file-storage-path=TEMP_FILE -database-dsn='postgres://postgres:password@localhost:15432/praktikum?sslmode=disable'