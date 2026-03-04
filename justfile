default:
    just --list

clean:
    go clean ./...
    rm -rf ./bin

build:
    mkdir -p ./bin
    go build -o ./bin/nacos-cli .

install:
    go install .

test:
    go test ./...
