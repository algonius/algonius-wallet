package integration

import "github.com/mark3labs/mcp-go/mcp"

// getTextContent extracts text content from MCP tool result
func getTextContent(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return ""
	}
	
	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		return ""
	}
	
	return textContent.Text
}