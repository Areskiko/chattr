package service

import (
	"context"

	external "github.com/areskiko/thatch/proto/inter"
)

type externalServer struct {
	external.UnimplementedExternalServiceServer
}

// Send implements external.ExternalServer.
func (*externalServer) Send(context.Context, *external.Message) (*external.Empty, error) {
	panic("unimplemented")
}

// mustEmbedUnimplementedExternalServer implements external.ExternalServer.
func (*externalServer) mustEmbedUnimplementedExternalServer() {
	panic("unimplemented")
}

