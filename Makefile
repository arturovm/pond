.PHONY: app
app: bin/pond
	mkdir -p bin/.pond

bin/pond: $(shell find . -path '**/*.go')
	go build -o bin/pond github.com/arturovm/pond/cmd/pond

.PHONY: test
test:
	go test ./...

.PHONY: run
run: bin/pond
	./bin/pond --debug

.PHONY: clean
clean:
	rm -rf bin
