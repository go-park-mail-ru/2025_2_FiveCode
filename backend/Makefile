.PHONY: test test-coverage run

test:
	go test ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	@grep -v "/mock/" coverage.out > coverage.out.tmp && mv coverage.out.tmp coverage.out
	go tool cover -func=coverage.out

run:
	go run main.go

.PHONY: proto
proto:
	protoc --proto_path=proto \
	       --go_out=gen/go --go_opt=paths=source_relative \
	       --go-grpc_out=gen/go --go-grpc_opt=paths=source_relative \
	       proto/auth/auth.proto proto/user/user.proto
