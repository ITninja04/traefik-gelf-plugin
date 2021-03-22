.PHONY: lint test vendor clean

export GO111MODULE=on

default: test

test:
	go test -v -cover ./...

test-report:
    go test -covermode=count -coverprofile=count.out .
    go tool cover -html=count.out
vendor:
	go mod vendor

clean:
	rm -rf ./vendor