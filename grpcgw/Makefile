.PHONY: generate_openapi

generate_openapi:
	@if ! command -v 'oapi-codegen' &> /dev/null; then \
		echo "Please install oapi-codegen!"; exit 1; \
	fi;
	@mkdir -p internal/types
	@oapi-codegen -package types \
		-generate types,gin \
		../openapi/v1/api.yaml > internal/types/api.gen.go
