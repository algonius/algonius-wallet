// Package toolutils provides utility functions for MCP tools.
package toolutils

import (
	"encoding/json"
	"fmt"

	"github.com/algonius/algonius-wallet/native/pkg/errors"
	"github.com/mark3labs/mcp-go/mcp"
)

// FormatErrorResult formats an error into a standardized tool result
func FormatErrorResult(err *errors.Error) *mcp.CallToolResult {
	// Convert error to JSON
	errorJSON, _ := json.Marshal(map[string]interface{}{
		"error": err,
	})

	// Create a tool result with both markdown text and JSON data
	toolResult := mcp.NewToolResultText(fmt.Sprintf("### Error\n\n- **Code**: `%s`\n- **Message**: %s\n- **Details**: %s\n- **Suggestion**: %s", 
		err.Code, err.Message, err.Details, err.Suggestion))
	
	// Add metadata with the raw JSON result
	if toolResult.Meta == nil {
		toolResult.Meta = make(map[string]any)
	}
	toolResult.Meta["error"] = string(errorJSON)
	toolResult.IsError = true
	
	return toolResult
}