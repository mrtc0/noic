build:
	go build -o noic .

install:
	cp ./noic /usr/local/bin/noic

test:
	go test ./...

create:
	sudo ./noic --debug --log tmp/log.json create test -b ./tmp/oci

start:
	sudo ./noic --debug --log tmp/log.json start test
	ps aux | grep sleep

debug:
	sudo strace -v -f -s 150 -p $(shell pidof noic)

delete:
	sudo ./noic --log tmp/log.json delete test
