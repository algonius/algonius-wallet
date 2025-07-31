package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/algonius/algonius-wallet/native/pkg/config"
	"github.com/algonius/algonius-wallet/native/pkg/dex"
	"github.com/algonius/algonius-wallet/native/pkg/dex/providers"
	"github.com/algonius/algonius-wallet/native/pkg/event"
	"github.com/algonius/algonius-wallet/native/pkg/logger"
	"github.com/algonius/algonius-wallet/native/pkg/mcp"
	"github.com/algonius/algonius-wallet/native/pkg/mcp/resources"
	"github.com/algonius/algonius-wallet/native/pkg/mcp/tools"
	"github.com/algonius/algonius-wallet/native/pkg/messaging"
	"github.com/algonius/algonius-wallet/native/pkg/messaging/handlers"
	"github.com/algonius/algonius-wallet/native/pkg/process"
	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/algonius/algonius-wallet/native/pkg/wallet/chain"
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

// setupUnifiedMCPServer creates a unified HTTP server supporting multiple MCP transport protocols
func setupUnifiedMCPServer(mcpServer *server.MCPServer, port string) *http.Server {
	mux := http.NewServeMux()
	
	// Streamable HTTP - compatible with existing clients
	streamableServer := server.NewStreamableHTTPServer(mcpServer, 
		server.WithEndpointPath("/mcp"))
	mux.Handle("/mcp", streamableServer)
	
	// Pure SSE - compatible with Cline and other SSE-only clients
	sseServer := server.NewSSEServer(mcpServer,
		server.WithStaticBasePath("/mcp"),
		server.WithSSEEndpoint("/sse"),
		server.WithMessageEndpoint("/message"),
		server.WithUseFullURLForMessageEndpoint(false)) // Use relative paths
	mux.Handle("/mcp/sse", sseServer.SSEHandler())
	mux.Handle("/mcp/message", sseServer.MessageHandler())
	
	return &http.Server{
		Addr:    port,
		Handler: mux,
	}
}

func main() {
	// Define command line flags
	killFlag := flag.Bool("kill", false, "Kill any existing instances of the native host")
	flag.Parse()
	
	// If kill flag is set, kill existing instances and exit
	if *killFlag {
		if err := process.KillExistingProcess(); err != nil {
			os.Stderr.WriteString("Failed to kill existing process: " + err.Error() + "\n")
			os.Exit(1)
		}
		os.Stderr.WriteString("Successfully killed existing instance (if any)\n")
		os.Exit(0)
	}
	
	// Skip process locking if ALGONIUS_WALLET_HOME is set (indicating test/isolated environment)
	isIsolatedEnvironment := os.Getenv("ALGONIUS_WALLET_HOME") != ""
	
	if !isIsolatedEnvironment {
		// Try to acquire PID file lock to prevent multiple instances
		locked, err := process.LockPIDFile()
		if err != nil {
			os.Stderr.WriteString("Failed to acquire PID file lock: " + err.Error() + "\n")
			os.Exit(1)
		}
		
		if !locked {
			os.Stderr.WriteString("Another instance of Algonius Native Host is already running\n")
			os.Exit(1)
		}
	}
	
	// Ensure we unlock the PID file when the program exits (only if not in isolated environment)
	defer func() {
		if !isIsolatedEnvironment {
			if err := process.UnlockPIDFile(); err != nil {
				// Log error but don't fail the program
				os.Stderr.WriteString("Failed to unlock PID file: " + err.Error() + "\n")
			}
		}
	}()

	// Initialize project logger
	logr, err := logger.NewLogger("main")
	if err != nil {
		os.Stderr.WriteString("Failed to initialize logger: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Extract zap logger from the wrapper for configuration loading
	zapLogger := logr.(*logger.ZapLogger).Logger

	// Load configuration 
	appConfig, err := config.LoadConfigWithFallback(zapLogger)
	if err != nil {
		zapLogger.Error("Failed to load configuration", zap.Error(err))
		os.Exit(1)
	}

	// Initialize DEX aggregator (placeholder for now)
	var dexAggregator dex.IDEXAggregator // nil for now, can be initialized later

	// Initialize global host state
	hostState := &HostState{
		startTime:  0, // will be set on init
		runMode:    "",
		version:    "0.1.0",
		shutdownCh: make(chan struct{}),
	}

	// Create shared wallet manager with configuration
	walletManager := wallet.NewWalletManagerWithConfig(appConfig, dexAggregator, zapLogger)
	
	// Create EventBroadcaster for real-time events to AI Agents
	eventBroadcaster := event.NewEventBroadcaster(zapLogger)

	logr.Info("Starting Algonius Native Host with both Native Messaging and HTTP/MCP servers")

	// Initialize Native Messaging for browser extension communication
	nm, err := messaging.NewNativeMessaging(messaging.NativeMessagingConfig{
		Logger: logr,
	})
	if err != nil {
		logr.Error("Failed to initialize native messaging", zap.Error(err))
		os.Exit(1)
	}

	// Register wallet RPC methods (only available via Native Messaging)
	nm.RegisterRpcMethod("import_wallet", handlers.CreateImportWalletHandler(walletManager))
	nm.RegisterRpcMethod("create_wallet", handlers.CreateCreateWalletHandler(walletManager))
	nm.RegisterRpcMethod("unlock_wallet", handlers.CreateUnlockWalletHandler(walletManager))
	nm.RegisterRpcMethod("lock_wallet", handlers.CreateLockWalletHandler(walletManager))
	nm.RegisterRpcMethod("wallet_status", handlers.CreateWalletStatusHandler(walletManager))
	nm.RegisterRpcMethod("web3_request", handlers.CreateWeb3RequestHandler(walletManager, eventBroadcaster))

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

	approveTransactionTool := tools.NewApproveTransactionTool(walletManager, eventBroadcaster, zapLogger)
	mcp.RegisterTool(s, approveTransactionTool)

	// Create DEX aggregator with OKX and Direct providers
	dexAggregator = dex.NewDEXAggregator(zapLogger)
	
	// Register Direct provider for backward compatibility
	directProvider := providers.NewDirectProvider(zapLogger)
	if err := dexAggregator.RegisterProvider(directProvider); err != nil {
		logr.Error("Failed to register direct provider", zap.Error(err))
	}
	
	// Check if we're in mock mode (for testing)
	if os.Getenv("DEX_MOCK_MODE") == "true" {
		// Register mock providers for testing
		mockProvider := providers.NewMockProvider(providers.MockConfig{
			Name: "MockOKX",
			SupportedChains: []string{"1", "56", "501"},
		}, zapLogger)
		if err := dexAggregator.RegisterProvider(mockProvider); err != nil {
			logr.Error("Failed to register mock provider", zap.Error(err))
		} else {
			logr.Info("Mock DEX provider registered for testing")
		}
	} else {
		// Register OKX provider if configured
		okxConfig := providers.OKXConfig{
			APIKey:     os.Getenv("OKX_API_KEY"),
			SecretKey:  os.Getenv("OKX_SECRET_KEY"),
			Passphrase: os.Getenv("OKX_PASSPHRASE"),
		}
		if okxConfig.APIKey != "" && okxConfig.SecretKey != "" && okxConfig.Passphrase != "" {
			okxProvider := providers.NewOKXProvider(okxConfig, zapLogger)
			if err := dexAggregator.RegisterProvider(okxProvider); err != nil {
				logr.Error("Failed to register OKX provider", zap.Error(err))
			} else {
				logr.Info("OKX DEX provider registered successfully")
			}
		} else {
			logr.Info("OKX credentials not provided, skipping OKX provider registration")
		}
	}
	
	swapTokensToolNew := tools.NewSwapTokensToolWithAggregator(dexAggregator, zapLogger)
	swapTokensToolNew.Register(s)

	getPendingTransactionsTool := tools.NewGetPendingTransactionsTool(walletManager)
	mcp.RegisterTool(s, getPendingTransactionsTool)

	getTransactionHistoryTool := tools.NewGetTransactionHistoryTool(walletManager)
	mcp.RegisterTool(s, getTransactionHistoryTool)

	// Create chain factory for simulation tools
	chainFactory := chain.NewChainFactory()

	// Register simulation tools
	simulateTransactionTool := tools.NewSimulateTransactionTool(walletManager, chainFactory)
	mcp.RegisterTool(s, simulateTransactionTool)

	// Start unified MCP server with multiple transport protocols
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		port := os.Getenv("SSE_PORT")
		if port == "" {
			port = ":9444"
		}
		unifiedServer := setupUnifiedMCPServer(s, port)
		logr.Info("Starting unified MCP server", 
			zap.String("port", port),
			zap.Strings("endpoints", []string{"/mcp", "/mcp/sse", "/mcp/message"}))
		if err := unifiedServer.ListenAndServe(); err != nil {
			logr.Error("Unified MCP Server error", zap.Error(err))
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