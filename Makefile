.PHONY: test
test: clean
	mkdir -p tmp
	go test -v ./...

.PHONY: build
build: clean test
	mkdir -p build
	go build -o build/gpxchart github.com/tkrajina/gpxchart/cmd/gpxchart/...
	@echo "Binary saved to build/gpxchart"

.PHONY: clean
clean:
	-rm tmp/tmp_chart*
	-rm build/*

.PHONY: generate
generate:
	go run gen.go

.PHONY: install
install:
	go install github.com/tkrajina/gpxchart/cmd/gpxchart/...

.PHONY: examples
examples: install
	python make_examples.py

.PHONY: lint
lint:
	golongfuncs -ignore ".*_generated.go"
	go vet ./...
	golangci-lint run