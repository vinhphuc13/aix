BINARY := aix
GOFLAGS := -ldflags="-s -w"

.PHONY: build install test clean

build:
	go build $(GOFLAGS) -o $(BINARY) .

install:
	go install $(GOFLAGS) .

test:
	go test ./...

clean:
	rm -f $(BINARY)

# Quick smoke test
smoke: install
	@echo "=== smoke test ==="
	@mkdir -p /tmp/aix-smoke && cd /tmp/aix-smoke && \
		aix start "smoke-test" --goal "verify build" && \
		aix add task "check status" && \
		aix add decision "use JSON" --rationale "simplicity" && \
		aix status && \
		aix checkpoint -m "smoke done" && \
		aix list && \
		echo '{"hook_event_name":"UserPromptSubmit","cwd":"/tmp/aix-smoke","prompt":"hello"}' | aix hook prompt | head -c 200 && \
		echo "" && \
		echo "=== smoke test passed ==="
	@rm -rf /tmp/aix-smoke/.aix
