package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type IResource interface {
	GetMeta() mcp.Resource
	GetHandler() server.ResourceHandlerFunc
}

type ITool interface {
	GetMeta() mcp.Tool
	GetHandler() server.ToolHandlerFunc
}

func RegisterResource(s *server.MCPServer, res IResource) {
	s.AddResource(res.GetMeta(), res.GetHandler())
}

func RegisterTool(s *server.MCPServer, tool ITool) {
	s.AddTool(tool.GetMeta(), tool.GetHandler())
}
