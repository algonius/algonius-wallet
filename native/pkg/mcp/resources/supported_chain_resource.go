// Package resources provides MCP resource implementations for the Algonius Native Host.
package resources

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// SupportedChainsResource implements the IResource interface for the "supported_chains" MCP resource.
// It returns the list of supported blockchain networks as a JSON array.
type SupportedChainsResource struct {
	Chains []string
}

// NewSupportedChainsResource creates a SupportedChainsResource with the default supported chains.
func NewSupportedChainsResource() *SupportedChainsResource {
	return &SupportedChainsResource{
		Chains: []string{"ethereum", "bsc", "solana"},
	}
}

// GetMeta returns the MCP resource definition for supported chains.
func (r *SupportedChainsResource) GetMeta() mcp.Resource {
	return mcp.NewResource(
		"chains://supported",
		"Supported Chains",
		mcp.WithResourceDescription("List of supported blockchain networks"),
		mcp.WithMIMEType("application/json"),
	)
}

// GetHandler returns the handler function for the supported chains resource.
// The handler marshals the supported chains list to JSON and returns it as resource contents.
func (r *SupportedChainsResource) GetHandler() server.ResourceHandlerFunc {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		data, err := json.Marshal(r.Chains)
		if err != nil {
			return nil, err
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "chains://supported",
				MIMEType: "application/json",
				Text:     string(data),
			},
		}, nil
	}
}
