# Protocol Buffers compilation tools
PROTOC := protoc
PROTOC_GEN_GO := protoc-gen-go
PROTOC_GEN_GO_GRPC := protoc-gen-go-grpc

# Directories
ROOT_DIR := $(shell cd ../../ && pwd)
PROTO_DIR := $(ROOT_DIR)/pkg/pb/protos
OUT_DIR := $(ROOT_DIR)/pkg/pb/gen

# Find all .proto files
PROTO_FILES := $(shell find $(PROTO_DIR) -name "*.proto")

.PHONY: proto_build proto_clean

# Main target to build all proto files
proto_build: proto_clean
	@cd $(ROOT_DIR) && \
	for proto in $(PROTO_FILES); do \
		$(PROTOC) --experimental_allow_proto3_optional \
			--go_out=$(OUT_DIR) --go-grpc_out=$(OUT_DIR) \
			--proto_path=$(ROOT_DIR) $$proto; \
		echo "Compiled $$proto"; \
	done


# Clean generated files
proto_clean:
	rm -rf $(OUT_DIR)/*

# Help target
help:
	@echo "Available targets:"
	@echo "  proto_build  - Compile all .proto files"
	@echo "  proto_clean  - Remove all generated files"
	@echo "  help         - Show this help message"

# Default target
.DEFAULT_GOAL := proto_build

