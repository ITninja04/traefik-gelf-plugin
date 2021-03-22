.PHONY: lint test vendor clean

export GO111MODULE=on

default: test

test:
	go test -v -cover ./...

vendor:
	go mod vendor

clean:
	rm -rf ./vendor