GO        ?= go
LINTER    ?= golangci-lint
ALIGNER   ?= betteralign
BENCH_REF ?= testdata/bench_baseline.txt

.PHONY: generate generate-check test test-race test-pure \
	bench verify vet fmt fmt-check lint align align-fix tidy download \
	check tools release-notes

check: fmt-check generate-check vet lint align test test-pure

generate:
	cd internal/simd/asmgen && GOWORK=off $(GO) run . \
		-out ../filters_amd64.s -stubs ../filters_stub_amd64.go -pkg simd
	gofmt -w internal/simd/filters_stub_amd64.go

generate-check: generate
	git diff --exit-code -- internal/simd

fmt:
	gofmt -w .

fmt-check:
	@gofmt -l . | tee /dev/stderr | read; \
	if [ $$? -eq 0 ]; then \
		echo "gofmt: files need formatting"; \
		exit 1; \
	fi

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

test-pure:
	$(GO) test -tags purego ./...

bench:
	@tmp=$$(mktemp); \
	$(GO) test -run=^$$ -bench 'Benchmark' -benchmem -count=6 | tee "$$tmp"; \
	if [ -f "$(BENCH_REF)" ]; then \
		benchstat "$(BENCH_REF)" "$$tmp"; \
	else \
		cp "$$tmp" "$(BENCH_REF)" && echo "Baseline saved to $(BENCH_REF)"; \
	fi; \
	rm -f "$$tmp"

verify:
	$(GO) mod verify

tidy:
	$(GO) mod tidy
	$(GO) -C ./internal/simd/asmgen mod tidy

download:
	$(GO) mod download
	$(GO) -C ./internal/simd/asmgen mod download

lint:
	$(LINTER) run ./...

align:
	$(ALIGNER) ./...

align-fix:
	-$(ALIGNER) -apply ./...
	$(ALIGNER) ./...

tools:
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	$(GO) install github.com/dkorunic/betteralign/cmd/betteralign@latest
	$(GO) install golang.org/x/perf/cmd/benchstat@latest

release-notes:
	@awk '\
	/^<!--/,/^-->/ { next } \
	/^## \[[0-9]+\.[0-9]+\.[0-9]+\]/ { if (found) exit; found=1; next } found { print } \
	' CHANGELOG.md
