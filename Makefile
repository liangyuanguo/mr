.PHONY: build
build:
	mkdir -p bin
	go build -o bin/mr-server ./cmd/server
	go build -o bin/mr-client ./cmd/client

.PHONY: clean
clean:
	rm -rf bin