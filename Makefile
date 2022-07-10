all: deps format check build
build:
	go build -x -v -a -ldflags "-s -w -X config.sha1ver=$(gitrev) -X config.buildTime=$(now)"
clean:
	rm -f tcp.ac
deps:
	go mod tidy -v
run:
	go run ./
check:
	go vet ./...
format:
	gofmt -s -l -w *.go config/*.go
