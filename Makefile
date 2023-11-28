NAME := github.com/pavelborisofff/go-metrics
DSN := 'postgres://postgres:password@localhost:15432/praktikum?sslmode=disable'

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

.PHONY: autotests
autotests: autotests14

.PHONY: autotests1
autotests1:
	go build -o cmd/server/server cmd/server/main.go
	go build -o cmd/agent/agent cmd/agent/main.go
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration1$$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server

.PHONY: autotests2
autotests2: autotests1
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration2[AB]*$$ -source-path=. -agent-binary-path=cmd/agent/agent

.PHONY: autotests3
autotests3: autotests2
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration3[AB]*$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server

.PHONY: autotests4
autotests4: autotests3
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration4$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port="8080"

.PHONY: autotests5
autotests5: autotests4
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration5$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port="8080"

.PHONY: autotests6
autotests6: autotests5
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration6$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port="8080"

.PHONY: autotests7
autotests7: autotests6
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration7$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port="8080" -file-storage-path=TEMP_FILE

.PHONY: autotests8
autotests8: autotests7
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration8$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port="8080" -file-storage-path=TEMP_FILE

.PHONY: autotests9
autotests9: autotests8
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration9$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port="8080" -file-storage-path=/tmp/metrics-db.json -file-storage-path=TEMP_FILE

.PHONY: autotests10
autotests10: autotests9
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration10[AB]$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port="8080" -database-dsn=$(DSN) -file-storage-path=TEMP_FILE

.PHONY: autotests11
autotests11: autotests10
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration11$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port="8080" -database-dsn=$(DSN) -file-storage-path=TEMP_FILE

.PHONY: autotests12
autotests12: autotests11
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration12$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port="8080" -database-dsn=$(DSN) -file-storage-path=TEMP_FILE

.PHONY: autotests13
autotests13: autotests12
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration13$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port="8080" -database-dsn=$(DSN) -file-storage-path=TEMP_FILE

.PHONY: autotests14
autotests14: autotests13
	metricstest-darwin-arm64 -test.v -test.run=^TestIteration14$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port="8080" -database-dsn=$(DSN) -key="secret" -file-storage-path=TEMP_FILE