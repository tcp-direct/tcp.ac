all: deps format check build
build-smol: all smol
build:
	go build -trimpath -x -v -a -ldflags "-s -w"
smol:
	upx --ultra-brute ./tcp.ac
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
