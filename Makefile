.PHONY: protoc
protoc:
	cd proto && \
	protoc --go_out=../genproto/go \
		--go_opt=paths=source_relative \
		--go-grpc_out=../genproto/go \
		--go-grpc_opt=paths=source_relative \
		./**/*.proto

