.PHONY: build clean test run

BINARY=bin/status-line
LDFLAGS=-ldflags="-s -w"

build:
	@mkdir -p bin
	go build $(LDFLAGS) -o $(BINARY) ./cmd/statusline

clean:
	rm -rf bin/

test:
	@echo '{"model":{"display_name":"Opus 4.5"},"workspace":{"current_dir":"~/project"},"context_window":{"total_input_tokens":15000,"total_output_tokens":5000,"context_window_size":200000}}' | ./$(BINARY)

run: build test
