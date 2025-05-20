.PHONY: run install update clean

BINARY_NAME=todo-cli

# Default target
all: install

# Run the application
run:
	go run main.go

# Install the application
install:
	go install .

# Update dependencies
update:
	go get -u ./...
	go mod tidy

# Clean build artifacts (optional, but good practice)
clean:
	go clean -i -n -x ./...
	rm -f $(BINARY_NAME)
