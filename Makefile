BINARY  = envy
PREFIX ?= /usr/local/bin

.PHONY: build install test clean

build:
	go build -o $(BINARY) ./cmd/envy/

install: build
	cp $(BINARY) $(PREFIX)/$(BINARY)

test:
	go test -v ./...

clean:
	rm -f $(BINARY)
