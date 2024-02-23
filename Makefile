.PHONY: all clean


PROTO_SRC_DIR = proto
PROTO_SRC = $(wildcard $(PROTO_SRC_DIR)/*.proto)
PROTO_GO = $(patsubst $(PROTO_SRC_DIR)/%.proto, $(PROTO_SRC_DIR)/%.pb.go, $(PROTO_SRC))


all: $(PROTO_GO)

# Compile proto files
$(PROTO_SRC_DIR)/%.pb.go: $(PROTO_SRC_DIR)/%.proto
	protoc --go_out=. \
		--go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		$<

clean:
	rm -f $(PROTO_GO)
