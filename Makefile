TARGET=driplane
LDFLAGS="-s -w"

all: build

test:
	@go test ./...

build: clean
	@mkdir build
	go build -o build/driplane -v -ldflags=${LDFLAGS} cmd/driplane/main.go

clean:
	@rm -rf build