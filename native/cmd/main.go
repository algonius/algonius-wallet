package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"

	"github.com/algonius/algonius-wallet/native/pkg/logger"
	"github.com/algonius/algonius-wallet/native/pkg/mcp"
	"github.com/algonius/algonius-wallet/native/pkg/mcp/resources"
	"github.com/algonius/algonius-wallet/native/pkg/mcp/tools"
	"github.com/algonius/algonius-wallet/native/pkg/messaging"
	"github.com/algonius/algonius-wallet/native/pkg/messaging/handlers"
	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Initialize project logger
	logr, err := logger.NewLogger("main")
	if err != nil {
		os.Stderr.WriteString("Failed to initialize logger: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Create shared wallet manager for both MCP and Native Messaging
	walletManager := wallet.NewWalletManager()

	logr.Info("Starting Algonius Native Host with both Native Messaging and HTTP/MCP servers")

	// Initialize Native Messaging for browser extension communication
	nm, err := messaging.NewNativeMessaging(messaging.NativeMessagingConfig{
		Logger: logr,
	})
	if err != nil {
		logr.Error("Failed to initialize native messaging", zap.Error(err))
		os.Exit(1)
	}

	// Register import_wallet RPC method (only available via Native Messaging)
	nm.RegisterRpcMethod("import_wallet", handlers.CreateImportWalletHandler(walletManager))

	// Create MCP server for AI Agent communication
	s := server.NewMCPServer(
		"Algonius Native Host",
		"0.1.0",
	)

	// Register chains://supported resource
	mcp.RegisterResource(s, resources.NewSupportedChainsResource())

	// Register wallet_status resource
	mcp.RegisterResource(s, resources.NewWalletStatusResource(walletManager))

	// Register MCP tools (no import_wallet tool as per security requirements)
	createWalletTool := tools.NewCreateWalletTool(walletManager)
	mcp.RegisterTool(s, createWalletTool)

	getBalanceTool := tools.NewGetBalanceTool(walletManager)
	mcp.RegisterTool(s, getBalanceTool)

	sendTransactionTool := tools.NewSendTransactionTool(walletManager)
	mcp.RegisterTool(s, sendTransactionTool)

	confirmTransactionTool := tools.NewConfirmTransactionTool(walletManager)
	mcp.RegisterTool(s, confirmTransactionTool)

	swapTokensTool := tools.NewSwapTokensTool(walletManager)
	mcp.RegisterTool(s, swapTokensTool)

	// Start HTTP MCP server in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpServer := server.NewStreamableHTTPServer(s)
		port := os.Getenv("SSE_PORT")
		if port == "" {
			port = ":8080"
		}
		logr.Info("Starting MCP HTTP server", zap.String("port", port), zap.String("endpoint", "/mcp"))
		if err := httpServer.Start(port); err != nil {
			logr.Error("HTTP Server error", zap.Error(err))
			os.Exit(1)
		}
	}()

	// Start Native Messaging in a goroutine as well
	wg.Add(1)
	go func() {
		defer wg.Done()
		logr.Info("Starting Native Messaging server")
		if err := nm.Start(); err != nil {
			logr.Error("Failed to start native messaging", zap.Error(err))
			os.Exit(1)
		}

		// Keep the Native Messaging goroutine alive by waiting for interrupt signal
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		logr.Info("Native Messaging received shutdown signal")
	}()

	// Wait for both servers to finish
	wg.Wait()
	logr.Info("Algonius Native Host shutdown complete")
}
