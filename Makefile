.PHONY: proto clean build test

# Generate protobuf files
proto:
	@echo "Generating protobuf files..."
	@protoc -I pkg/proto --go_out=. --go-grpc_out=. pkg/proto/*.proto
	@echo "Protobuf files generated successfully."

LOG_FILE ?= logs/$(shell date +%Y%m%d_%H%M%S).log