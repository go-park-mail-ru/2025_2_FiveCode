.PHONY: test test-coverage run

test:
	go test ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	@grep -v "/mock/" coverage.out > coverage.out.tmp && mv coverage.out.tmp coverage.out
	go tool cover -func=coverage.out

run:
	go run main.go
