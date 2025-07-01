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
	@echo "Running tests..."
	go test -v ./tests/...

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
	@echo "Installing native host..."
	sudo cp $(BINARY_NAME) $(INSTALL_DIR)/
	sudo cp $(MANIFEST_NAME) $(MANIFEST_DIR)/
	sudo chmod 755 $(INSTALL_DIR)/$(BINARY_NAME)
	sudo chmod 644 $(MANIFEST_DIR)/$(MANIFEST_NAME)
	
	# Create configuration directory
	mkdir -p ~/.algonius-wallet
	chmod 700 ~/.algonius-wallet
	
	@echo "Native host installed successfully"

uninstall:
	@echo "Uninstalling native host..."
	sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	sudo rm -f $(MANIFEST_DIR)/$(MANIFEST_NAME)
	@echo "Native host uninstalled"

.PHONY: all build unit-test integration-test test run clean install uninstall lint
