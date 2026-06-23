// Package mcpserver builds the MCP server whose tools each wrap exactly one
// *store.Store method — the same method the corresponding kong CLI command
// calls. CLI command ⇄ MCP tool are two front doors to one function; no logic
// lives in either wrapper.
package mcpserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// New creates and returns an MCP server with all quilt tools, the always-on
// operating instructions, and the guide://·view:// resources and prompts that
// make the server self-documenting.
func New(version string) *mcp.Server {
	s := mcp.NewServer(
		&mcp.Implementation{Name: "quilt", Version: version},
		&mcp.ServerOptions{Instructions: instructions},
	)
	registerTools(s)
	registerDocs(s)
	return s
}

// RunStdio serves the MCP server over stdio (the default transport, used by
// Claude Desktop and the .dxt package).
func RunStdio(ctx context.Context, version string) error {
	return New(version).Run(ctx, &mcp.StdioTransport{})
}

// RunHTTP serves the MCP server over a Streamable HTTP listener at addr. Each
// session reuses the same server instance; the tools are transport-agnostic.
func RunHTTP(ctx context.Context, version, addr string) error {
	srv := New(version)
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return srv }, nil)
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		_ = httpSrv.Close()
	}()
	if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
