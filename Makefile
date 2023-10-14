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

