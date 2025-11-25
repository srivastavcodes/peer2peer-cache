.PHONY: run
run:
	echo "running application"
	go run .

.PHONY: test
test:
	echo "running tests"
	go test -race .

.PHONY: compile
compile:
	protoc p2pcachepb/v1/*.proto \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		--proto_path=.
