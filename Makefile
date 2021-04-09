buildDateTime = $(shell date '+%Y-%m-%d %H:%M:%S')
gitCommitCode = $(shell git rev-list --full-history --all --abbrev-commit --max-count 1)
goVersion = $(shell go version)

# build dev
dev:
	cd cmd && go build -o ../deployments/bareservice-go
	./deployments/bareservice-go -debug

release:
	cd cmd && gox -osarch="linux/arm" -ldflags "-X 'main.buildDateTime=$(buildDateTime)' -X 'main.gitCommitCode=$(gitCommitCode)' -X "main.goVersion=$(gitVersion)" -s -w" -output ../deployments/bareservice-go
	cd deployments && upx -9 bareservice-go && chmod +x bareservice-go