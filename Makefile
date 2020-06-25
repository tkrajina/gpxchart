.PHONY: test
test: clean
	go test -v ./...

.PHONY: clean
clean:
	-rm gpxcharts/tmp_chart*

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