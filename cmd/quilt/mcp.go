package main

import (
	"context"

	"github.com/bmayfi3ld/quilt/pkg/mcpserver"
)

type mcpCmd struct {
	HTTP string `name:"http" help:"serve over Streamable HTTP at this address (e.g. :8080) instead of stdio"`
}

func (c *mcpCmd) Run() error {
	ctx := context.Background()
	if c.HTTP != "" {
		return mcpserver.RunHTTP(ctx, version, c.HTTP)
	}
	return mcpserver.RunStdio(ctx, version)
}
