GO := CGO_ENABLED=0 go

.PHONY: test build site serve lint clean

test:
	$(GO) test ./...

lint:
	gofmt -l . && $(GO) vet ./...

build:
	$(GO) build -o pgbook-bin .

# Regenerate site/api from topics/*.md; run after editing any topic.
site:
	$(GO) run ./tools/gensite
	cp install.sh site/install.sh

# Preview pgbook.dev locally on http://127.0.0.1:8391
serve: site
	$(GO) run ./tools/serve

clean:
	rm -f pgbook-bin
	rm -rf dist
