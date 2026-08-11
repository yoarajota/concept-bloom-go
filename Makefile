.PHONY: quality test bench

quality:
	go vet ./...
	staticcheck ./...
	gitleaks detect --no-banner
	govulncheck ./...

test:
	go test ./... -v -count=1

bench:
	go test ./bench/... -bench=. -benchmem -count=5
