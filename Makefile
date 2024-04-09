all: clean local
clean:
	rm -rf out

local:
	mkdir out
	go build -o out/mistrecvd cmd/mistrecvd/main.go

