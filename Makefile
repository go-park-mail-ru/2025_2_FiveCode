.PHONY: test test-coverage run

test:
	go test ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	grep -v "/mock/" coverage.out | grep -v ".pb.go" > coverage.out.tmp
	mv coverage.out.tmp coverage.out
	go tool cover -func=coverage.out

run:
	go run main.go

.PHONY: proto
proto:
	protoc -I . \
		--go_out=. --go_opt=module=backend \
		--go-grpc_out=. --go-grpc_opt=module=backend \
		auth_service/proto/auth/v1/auth.proto \
		user_service/proto/user/v1/user.proto \
		notes_service/proto/note/v1/note.proto \
		notes_service/proto/block/v1/block.proto
