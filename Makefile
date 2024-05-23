all: clean mistwsrecvd mistpolld
clean:
	rm -rf out

mistwsrecvd:
	mkdir -p out
	go build -o out/mistwsrecvd cmd/mistwsrecvd/main.go

mistpolld:
	mkdir -p out
	go build -o out/mistpolld cmd/mistpolld/main.go

