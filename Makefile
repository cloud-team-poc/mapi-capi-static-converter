BIN_NAME=converter

default: build

help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''

build:  ## Compile the project
	go build -o bin/${BIN_NAME}

test: ## Test the project
	go test ./pkg/...

clean:
	rm -rf bin/
