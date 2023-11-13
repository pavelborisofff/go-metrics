NAME = github.com/pavelborisofff/go-metrics

.PHONY: go-init
go-init:
	go mod init $(NAME)

.PHONY: go-run-server
go-run-server:
	cd ./cmd/server && go run .

.PHONY: go-run-server-race
go-run-server-race:
	cd ./cmd/server && go run -race .

.PHONY: go-run-agent
go-run-agent:
	cd ./cmd/agent && go run .

.PHONY: go-run-agent-race
go-run-agent-race:
	cd ./cmd/agent && go run -race .

.PHONY: go-run-tests
go-run-tests:
	go test ./...

.PHONY: git-checkout
git-checkout:
	git checkout -b $(BRANCH)

.PHONY: go-run-autotests
go-run-autotests:
	go build -o cmd/server/server cmd/server/main.go
	go build -o cmd/agent/agent cmd/agent/main.go
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration${TEST_NUM}$$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=12345 -source-path=. -file-storage-path=TEMP_FILE -database-dsn='postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable'