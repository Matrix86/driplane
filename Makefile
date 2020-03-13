test:
	@go test ./...

build: clean
	@mkdir build
	go build -o build/driplane cmd/driplane/main.go

clean:
	@rm -rf build