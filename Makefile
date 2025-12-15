.PHONY: build clean test lint demo run

BINARY=bin/status-line
LDFLAGS=-ldflags="-s -w"

# Build the status-line binary
build:
	@mkdir -p bin
	go build $(LDFLAGS) -o $(BINARY) ./cmd/statusline

# Remove build artifacts
clean:
	rm -rf bin/

# Run Go unit tests
test:
	go test ./...

# Run ktn-linter on the codebase
lint:
	@command -v ktn-linter >/dev/null 2>&1 || { echo "Installing ktn-linter..."; go install github.com/kodflow/ktn-linter@latest; }
	ktn-linter lint ./...

# Demo the status line with sample input
demo: build
	@echo '{"model":{"display_name":"Opus 4.5"},"workspace":{"current_dir":"~/project"},"context_window":{"total_input_tokens":15000,"total_output_tokens":5000,"context_window_size":200000}}' | ./$(BINARY)

# Build and demo
run: build demo
