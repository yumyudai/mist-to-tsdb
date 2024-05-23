VER := $(shell git rev-parse HEAD | tr -d "\n")
all: clean mistwsrecvd mistpolld
clean:
	rm -rf out

container: mistpolld-container mistwsrecvd-container

mistwsrecvd:
	mkdir -p out
	go build -o out/mistwsrecvd cmd/mistwsrecvd/main.go

mistwsrecvd-container:
	docker build -t mistwsrecvd:$(VER) -f build/mistwsrecvd/Dockerfile .
	docker tag mistwsrecvd:$(VER) mistwsrecvd:latest

mistpolld:
	mkdir -p out
	go build -o out/mistpolld cmd/mistpolld/main.go

mistpolld-container:
	docker build -t mistpolld:$(VER) -f build/mistpolld/Dockerfile .
	docker tag mistpolld:$(VER) mistpolld:latest
