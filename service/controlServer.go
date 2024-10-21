package main

import (
	"context"
	"log/slog"
	"sync"

	internal "github.com/areskiko/thatch/proto/intra"
)

type controlServer struct {
	internal.UnimplementedControlServiceServer
	wg *sync.WaitGroup
}

// Kill implements intra.ControlServiceServer.
func (c *controlServer) Kill(context.Context, *internal.KillRequest) (*internal.KillResponse, error) {
	slog.Info("Received kill request")
	c.wg.Done()

	return &internal.KillResponse{}, nil
}

// Scan implements intra.ControlServiceServer.
func (c *controlServer) Scan(context.Context, *internal.ScanRequest) (*internal.ScanResponse, error) {
	slog.Info("Received scan request")
	go discover()

	return &internal.ScanResponse{}, nil
}
