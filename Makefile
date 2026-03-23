.PHONY: build run clean test vet fmt

build:
	go build -o awry ./cmd/awry

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w ./cmd ./internal ./pkg

run: build
	./awry

clean:
	rm -f awry
