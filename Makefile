all: test build

build:
	@echo ">> building ringpop"
	@GO111MODULE="on" GOOS="linux" GOARCH="amd64" go build -tags netgo -installsuffix netgo -o target/ringpop cmd/ringpop/main.go

build-backend:
	@echo ">> building backend"
	@GO111MODULE="on" GOOS="linux" GOARCH="amd64" go build -tags netgo -installsuffix netgo -o target/backend cmd/backend-example/main.go

test:
	@echo ">> making tests"
	@go test ./...

clean:
	@echo ">> removing build directory"
	@rm -rf target

fmt:
	@echo ">> formatting source"
	@find . -type f -iname '*.go' -not -path './vendor/*' -not -iname '*pb.go' | xargs -L 1 go fmt

imports:
	@echo ">> fixing source imports"
	@find . -type f -iname '*.go' -not -path './vendor/*' -not -iname '*pb.go' | xargs -L 1 goimports -w

lint:
	@echo ">> linting source"
	@find . -type f -iname '*.go' -not -path './vendor/*' -not -iname '*pb.go' | xargs -L 1 golint
