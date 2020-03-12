test:
	@go test ./...

binary: clean
	@mkdir build
	go build -o build/driplane cmd/driplane/main.go

clean:
	@rm -rf build