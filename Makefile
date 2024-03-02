.PHONY: all check clean tui service

BUILD_DIR = build

PROTO_SRC_DIR = proto
PROTO_SRC = $(wildcard $(PROTO_SRC_DIR)/*/*.proto)
PROTO_GO = $(patsubst $(PROTO_SRC_DIR)/%.proto, $(PROTO_SRC_DIR)/%.pb.go, $(PROTO_SRC))
GRPC_GO = $(patsubst $(PROTO_SRC_DIR)/%.proto, $(PROTO_SRC_DIR)/%_grpc.pb.go, $(PROTO_SRC))

INTERNAL = $(PROTO_SRC_DIR)/intra/internal.pb.go $(PROTO_SRC_DIR)/intra/internal_grpc.pb.go
EXTERNAL = $(PROTO_SRC_DIR)/inter/external.pb.go $(PROTO_SRC_DIR)/inter/external_grpc.pb.go

TUI_GO = $(wildcard .tui/*.go)
SERVICE_GO = $(wildcard .service/*.go)


all: tui service

tui: $(BUILD_DIR)/tui
service: $(BUILD_DIR)/service

$(BUILD_DIR)/tui: $(TUI_GO) $(INTERNAL)
	go build -o $(BUILD_DIR)/tui ./tui

$(BUILD_DIR)/service: $(SERVICE_GO) $(INTERNAL) $(EXTERNAL)
	go build -o $(BUILD_DIR)/service ./service

$(PROTO_SRC_DIR)/%.pb.go: $(PROTO_SRC_DIR)/%.proto
	protoc --go_out=$(PROTO_SRC_DIR) --go_opt=paths=source_relative --proto_path=$(PROTO_SRC_DIR) $<

$(PROTO_SRC_DIR)/%_grpc.pb.go: $(PROTO_SRC_DIR)/%.proto
	protoc --go-grpc_out=$(PROTO_SRC_DIR) --go-grpc_opt=paths=source_relative  --proto_path=$(PROTO_SRC_DIR) $<

check:
	go vet ./tui
	go vet ./service

clean:
	@ rm -rf $(BUILD_DIR)
	@ rm -f $(PROTO_GO)
	@ rm -f $(GRPC_GO)

