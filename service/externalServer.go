package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strings"

	external "github.com/areskiko/thatch/proto/inter"
	internal "github.com/areskiko/thatch/proto/intra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type externalServer struct {
	external.UnimplementedExternalServiceServer
}

func verifySignature(sig string, msg string, auth string) bool {
	return true
}

// Send implements external.ExternalServer.
func (*externalServer) Send(ctx context.Context, request *external.SendRequest) (*external.SendResponse, error) {
	if !verifySignature(request.GetSignature(), request.GetMessage(), request.GetAuthentication()) {
		return nil, errors.New("Invalid signature for identity")
	}

	chat := chats[request.GetAuthentication()]
	if chat == nil {
		return nil, errors.New("No chat for identity")
	}

	if chat.Messages == nil {
		chat.Messages = make([]*internal.Message, 1)
	}

	chat.Messages = append(chat.Messages, &internal.Message{Sender: request.Authentication, Text: request.GetMessage()})

	return &external.SendResponse{}, nil
}

func (*externalServer) Handshake(ctx context.Context, request *external.HandshakeRequest) (*external.HandshakeResponse, error) {
	slog.Info("Got handshake request")
	addr, err := net.ResolveTCPAddr("tcp", request.GetAddress())
	if err != nil {
		return nil, err
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return net.Dial("tcp", addr)
	}

	opts = append(opts, grpc.WithContextDialer(dialer))
	conn, err := grpc.Dial(addr.String(), opts...)
	if err != nil {
		slog.Debug("Failed to connect to server: %v\n", err)
		return nil, err
	}

	client := external.NewExternalServiceClient(conn)
	peers[request.GetAuthentication()] = &Peer{address: addr, authentication: request.GetAuthentication(), client: client}
	users[request.GetName()] = &internal.User{Name: strings.Split(request.GetName(), "#")[0], Id: request.GetAuthentication()}

	return &external.HandshakeResponse{Name: *cli.Username, Authentication: *cli.Username}, nil
}
