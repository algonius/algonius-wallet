# Algonius Wallet Native Host Makefile

# Configuration
BINARY_NAME=algonius-wallet-host
MANIFEST_NAME=com.algonius.wallet.json
PKG_PATH=./cmd
INSTALL_DIR=/usr/local/bin
MANIFEST_DIR=/etc/opt/chrome/native-messaging-hosts

# Build targets
all: build

build:
	@echo "Building native host..."
	go build -o ./bin/$(BINARY_NAME) $(PKG_PATH)/main.go

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo >&2 "golangci-lint is not installed. See https://golangci-lint.run/usage/install/"; exit 1; }
	@golangci-lint run --config .golangci.yml ./...

unit-test:
	@echo "Running tests..."
	go test -v $(PKG_PATH)/... ./pkg/...

integration-test:
	@echo "Running integration tests..."
	@echo "Building binary first..."
	@$(MAKE) build
	@echo "Running core integration tests..."
	DEX_MOCK_MODE=true go test -v -timeout=3m -run "TestChainsSupportedResource$$|TestCreateWalletTool$$|TestGetBalanceTool$$|TestWalletStatusResource$$|TestApproveTransactionTool$$" ./tests/integration/

integration-test-extended:
	@echo "Running extended integration tests..."
	@echo "Building binary first..."
	@$(MAKE) build
	@echo "Running additional working integration tests..."
	DEX_MOCK_MODE=true go test -v -timeout=5m -run "TestChains|TestCreate|TestGetBalance|TestWalletStatus|TestGetPendingTransactions|TestSendTransaction|TestSwap|TestImport|TestSSE" ./tests/integration/

integration-test-all:
	@echo "Running ALL integration tests..."
	@echo "Building binary first..."
	@$(MAKE) build
	@echo "Running all tests with extended timeout..."
	DEX_MOCK_MODE=true go test -v -timeout=10m ./tests/integration/...

test: unit-test integration-test

run:
	@echo "Running native host..."
	go run ./cmd/main.go

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	rm -f $(MANIFEST_DIR)/$(MANIFEST_NAME)

install: build
	@echo "Killing any existing instances..."
	@-./bin/$(BINARY_NAME) -kill 2>/dev/null || true
	@echo "Installing native host to $(HOME)/.algonius-wallet/bin/algonius-wallet-host ..."
	mkdir -p $(HOME)/.algonius-wallet/bin
	cp ./bin/$(BINARY_NAME) $(HOME)/.algonius-wallet/bin/algonius-wallet-host
	chmod 755 $(HOME)/.algonius-wallet/bin/algonius-wallet-host

	# Install manifest file
	@echo "Installing manifest file..."
	mkdir -p $(HOME)/.algonius-wallet
	# Create manifest with correct path
	sed "s|/Users/\$(USER)|$(HOME)|g" ./$(MANIFEST_NAME) > $(HOME)/.algonius-wallet/$(MANIFEST_NAME)
	
	# Optionally create configuration directory
	mkdir -p $(HOME)/.algonius-wallet
	chmod 700 $(HOME)/.algonius-wallet

	@echo "Native host installed successfully at $(HOME)/.algonius-wallet/bin/algonius-wallet-host"

uninstall:
	@echo "Uninstalling native host..."
	sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	sudo rm -f $(MANIFEST_DIR)/$(MANIFEST_NAME)
	@echo "Native host uninstalled"

.PHONY: all build unit-test integration-test integration-test-extended integration-test-all test run clean install uninstall lint kill help

kill:
	@echo "Killing any running instances of algonius-wallet-host..."
	@-pkill -f algonius-wallet-host 2>/dev/null || true
	@echo "Done."

help:
	@echo "Available targets:"
	@echo "  build                     - Build the native host binary"
	@echo "  unit-test                 - Run unit tests"
	@echo "  integration-test          - Run core integration tests"
	@echo "  integration-test-extended - Run extended integration tests"
	@echo "  integration-test-all      - Run ALL integration tests"
	@echo "  test                      - Run both unit and core integration tests"
	@echo "  lint                      - Run linter"
	@echo "  run                       - Run the native host"
	@echo "  install                   - Install the native host"
	@echo "  uninstall                 - Uninstall the native host"
	@echo "  kill                      - Kill any running instances"
	@echo "  clean                     - Clean build artifacts"
	@echo "  help                      - Show this help"
