NAME = github.com/pavelborisofff/go-metrics

.PHONY: go-init
go-init:
	go mod init $(NAME)

.PHONY: go-run-server
go-run-server:
	cd ./cmd/server && go run .

.PHONY: go-run-agent
go-run-agent:
	cd ./cmd/agent && go run .

.PHONY: git-checkout
git-checkout:
	git checkout -b $(BRANCH)

