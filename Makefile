BINARY := datasetlint

.PHONY: build test example release-check snapshot

build:
	mkdir -p dist
	go build -o dist/$(BINARY) .

test:
	go test ./...

example:
	go run . scan -train examples/train.jsonl -eval examples/eval.jsonl

release-check:
	goreleaser check

snapshot:
	goreleaser release --snapshot --clean

