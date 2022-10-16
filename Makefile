build:
	go build -o noic .

install:
	cp ./noic /usr/local/bin/noic

test:
	go test ./...
