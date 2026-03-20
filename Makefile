.PHONY: build run clean

build:
	go build -o awry ./cmd/awry

run: build
	./awry

clean:
	rm -f awry
