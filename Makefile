build:
	CGO_ENABLED=0 go build -o estimate ./cmd/estimate/

run: build
	./estimate

test:
	go test ./...

clean:
	rm -f estimate

.PHONY: build run test clean
