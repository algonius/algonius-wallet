package main

import (
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

type HostState struct {
	startTime  int64
	runMode    string
	version    string
	shutdownCh chan struct{}
}

func makeTimestamp() int64 {
	return time.Now().Unix()
}

func main() {
	// Initialize project logger
	logr, err := logger.NewLogger("main")
	if err != nil {
		os.Stderr.WriteString("Failed to initialize logger: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Initialize global host state
	hostState := &HostState{
		startTime:  0, // will be set on init
		runMode:    "",
		version:    "0.1.0",
		shutdownCh: make(chan struct{}),
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

	// Register init, status, shutdown RPC methods
	nm.RegisterRpcMethod("init", func(req messaging.RpcRequest) (messaging.RpcResponse, error) {
		// Parse params: { runMode: string, port?: number, logLevel?: string }
		type InitParams struct {
			RunMode  string `json:"runMode"`
			Port     int    `json:"port,omitempty"`
			LogLevel string `json:"logLevel,omitempty"`
		}
		var params InitParams
		if req.Params != nil {
			if err := json.Unmarshal(req.Params, &params); err != nil {
				return messaging.RpcResponse{
					ID: req.ID,
					Error: &messaging.ErrorInfo{
						Code:    -32602,
						Message: "Invalid params: " + err.Error(),
					},
				}, nil
			}
		}
		if params.RunMode == "" {
			return messaging.RpcResponse{
				ID: req.ID,
				Error: &messaging.ErrorInfo{
					Code:    -32602,
					Message: "runMode is required",
				},
			}, nil
		}
		hostState.runMode = params.RunMode
		hostState.startTime = makeTimestamp()
		// Optionally: set log level if needed
		result, _ := json.Marshal(map[string]string{"status": "initialized"})
		return messaging.RpcResponse{
			ID:     req.ID,
			Result: result,
		}, nil
	})
	nm.RegisterRpcMethod("status", func(req messaging.RpcRequest) (messaging.RpcResponse, error) {
		now := time.Now()
		startTime := time.Unix(hostState.startTime, 0)
		uptime := now.Sub(startTime)
		status := map[string]interface{}{
			"version":      hostState.version,
			"runMode":      hostState.runMode,
			"start_time":   startTime.Format(time.RFC3339),
			"current_time": now.Format(time.RFC3339),
			"uptime":       uptime.String(),
		}
		result, _ := json.Marshal(status)
		return messaging.RpcResponse{
			ID:     req.ID,
			Result: result,
		}, nil
	})
	nm.RegisterRpcMethod("shutdown", func(req messaging.RpcRequest) (messaging.RpcResponse, error) {
		select {
		case hostState.shutdownCh <- struct{}{}:
		default:
			// already shutting down or channel full
		}
		result, _ := json.Marshal(map[string]string{"status": "shutting_down"})
		return messaging.RpcResponse{
			ID:     req.ID,
			Result: result,
		}, nil
	})

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

	// Start HTTP MCP server with SSE support in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		
		// Create combined handler that supports both HTTP stream and SSE transports
		mcpHandler := mcp.CreateMCPHandler(s, logr, "Algonius Native Host", "0.1.0")
		
		// Set up HTTP server with multiple endpoints
		mux := http.NewServeMux()
		mux.HandleFunc("/mcp", mcpHandler)              // Main MCP endpoint with auto-detection
		mux.HandleFunc("/mcp/sse", mcpHandler)          // Explicit SSE endpoint
		mux.HandleFunc("/mcp/stream", mcpHandler)       // Explicit HTTP stream endpoint
		
		port := os.Getenv("SSE_PORT")
		if port == "" {
			port = ":9444"
		}
		
		logr.Info("Starting MCP server with dual transport support", 
			zap.String("port", port), 
			zap.String("endpoints", "/mcp, /mcp/sse, /mcp/stream"),
			zap.String("transports", "HTTP stream + SSE"))
		
		if err := http.ListenAndServe(port, mux); err != nil {
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

		// Wait for either shutdown RPC or OS signal
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		select {
		case <-hostState.shutdownCh:
			logr.Info("Native Messaging received shutdown RPC, delaying exit for 1s")
			time.Sleep(1 * time.Second)
		case <-c:
			logr.Info("Native Messaging received OS shutdown signal")
		}
	}()

	// Wait for both servers to finish
	wg.Wait()
	logr.Info("Algonius Native Host shutdown complete")
}
