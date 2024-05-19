all: clean local
clean:
	rm -rf out

local:
	mkdir out
	go build -o out/mistwsrecvd cmd/mistwsrecvd/main.go

