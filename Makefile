.PHONY: all clean

all: build

build:
	go build -o bin/scp scp.go

clean:
	rm -f bin/scp
