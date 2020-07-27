buildDateTime = $(shell date '+%Y-%m-%d %H:%M:%S')
gitCommitCode = $(shell git rev-list --full-history --all --abbrev-commit --max-count 1)
goVersion = $(shell go version)

# build dev
dev:
	cd cmd && go build -o ../deployments/baresip-go
	./deployments/baresip-go -debug

release:
	cd cmd && gox -osarch="linux/arm" -ldflags "-X 'main.buildDateTime=$(buildDateTime)' -X 'main.gitCommitCode=$(gitCommitCode)' -s -w" -output ../deployments/baresip-go
	cd deployments && upx -9 baresip-go && chmod +x baresip-go