.PHONY: build run test clean

build:
	go build -o glace .

run: build
	./glace

test:
	go test ./...

clean:
	rm -f glace
