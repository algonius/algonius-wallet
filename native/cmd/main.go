package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/algonius/algonius-wallet/native/pkg/events"
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
		_, _ = os.Stderr.WriteString("Failed to initialize logger: " + err.Error() + "\n")
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

	// Create event broadcaster for real-time notifications
	eventBroadcaster := events.NewEventBroadcaster(logr)

	// Get server type from environment variable
	serverType := os.Getenv("MCP_SERVER_TYPE")
	if serverType == "" {
		serverType = "http_stream" // Default to HTTP Stream
	}

	logr.Info("Starting Algonius Native Host with both Native Messaging and MCP servers",
		zap.String("mcp_server_type", serverType))

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

	// Register events stream resource with resource manager
	resourceManager := mcp.NewResourceManager(s)
	eventsStreamResource := resources.NewEventsStreamResource(eventBroadcaster, logr)
	eventsStreamResource.SetResourceManager(resourceManager)
	mcp.RegisterResource(s, eventsStreamResource)

	// Register SSE events resource for event type documentation
	sseEventsResource := resources.NewSSEEventsResource(eventBroadcaster, logr)
	mcp.RegisterResource(s, sseEventsResource)

	// Register MCP tools (no import_wallet tool as per security requirements)
	createWalletTool := tools.NewCreateWalletTool(walletManager)
	mcp.RegisterTool(s, createWalletTool)

	getBalanceTool := tools.NewGetBalanceTool(walletManager)
	mcp.RegisterTool(s, getBalanceTool)

	sendTransactionTool := tools.NewSendTransactionToolWithBroadcaster(walletManager, eventBroadcaster)
	mcp.RegisterTool(s, sendTransactionTool)

	confirmTransactionTool := tools.NewConfirmTransactionToolWithBroadcaster(walletManager, eventBroadcaster)
	mcp.RegisterTool(s, confirmTransactionTool)

	swapTokensTool := tools.NewSwapTokensTool(walletManager)
	mcp.RegisterTool(s, swapTokensTool)

	// Start MCP server based on environment variable
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		port := os.Getenv("MCP_PORT")
		if port == "" {
			port = "9444"
		}
		if port[0] != ':' {
			port = ":" + port
		}

		switch serverType {
		case "sse":
			logr.Info("Starting MCP SSE server", zap.String("port", port), zap.String("endpoint", "/sse"))
			sseServer := server.NewSSEServer(s)
			if err := sseServer.Start(port); err != nil {
				logr.Error("SSE Server error", zap.Error(err))
				os.Exit(1)
			}
		case "http_stream":
		default:
			logr.Info("Starting MCP HTTP Stream server", zap.String("port", port), zap.String("endpoint", "/mcp"))
			httpServer := server.NewStreamableHTTPServer(s)
			if err := httpServer.Start(port); err != nil {
				logr.Error("HTTP Stream Server error", zap.Error(err))
				os.Exit(1)
			}
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
