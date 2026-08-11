.PHONY: quality test bench

quality:
	go vet ./...
	gocyclo -over 15 .
	golangci-lint run
	gitleaks detect --no-banner

test:
	go test ./internal/... -v -count=1

bench:
	go test ./bench/... -bench=. -benchmem -count=5
