.PHONY: all depends test install clean

all: qrserve

depends:
	go get github.com/skip2/go-qrcode

qrserve: 
	go build

test:
	go test

install:
	go install

clean:
	go clean

docker:
	export GOOS=linux; export GOARCH=amd64; go build -o qrserve
	docker build . -t yltd/qrserve:latest
