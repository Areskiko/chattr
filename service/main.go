package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/netip"
	"os"
	"strings"
	"sync"

	"github.com/alecthomas/kong"
	external "github.com/areskiko/thatch/proto/inter"
	internal "github.com/areskiko/thatch/proto/intra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Peer struct {
	address        *net.TCPAddr
	authentication string
	client         external.ExternalServiceClient
}

var (
	chats map[string]*internal.Chat = make(map[string]*internal.Chat)
	users map[string]*internal.User = make(map[string]*internal.User)
	peers map[string]*Peer          = make(map[string]*Peer)
)

func reachOut(addr *net.TCPAddr) {

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return net.Dial("tcp", addr)
	}
	opts = append(opts, grpc.WithContextDialer(dialer))

	conn, err := grpc.Dial(addr.String(), opts...)
	if err != nil {
		slog.Debug("Failed to connect to server: %v\n", err)
		return
	}

	client := external.NewExternalServiceClient(conn)
	result, err := client.Handshake(context.Background(), &external.HandshakeRequest{Name: *cli.Username, Authentication: *cli.Username})
	if err != nil {
		slog.Debug("Failed to handshake with server: %v\n", err)
		return
	}

	peers[result.GetName()] = &Peer{address: addr, authentication: result.GetAuthentication(), client: client}
	users[result.GetName()] = &internal.User{Name: strings.Split(result.GetName(), "#")[0], Id: result.GetName()}
	slog.Info(fmt.Sprintf("Handshake with %s complete", addr))

}

func discover() {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		slog.Error("Could not create connection for discovery", err)
		return
	}

	defer conn.Close()

	local := conn.LocalAddr().(*net.UDPAddr)
	ip := local.IP
	cidr := fmt.Sprintf("%s/%d", ip, cli.SubnetMask)
	p, err := netip.ParsePrefix(cidr)
	if err != nil {
		slog.Error("Could not parse CIDR: %s", err)
	}

	p = p.Masked()
	slog.Debug(fmt.Sprintf("Scanning %s", p.String()))

	addr := p.Addr()
	for {
		if !p.Contains(addr) {
			break
		}

		ad, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", addr.String(), cli.ScanningPort))
		if err != nil {
			slog.Warn("Error resolving %s: %s", addr.String(), err)
			return
		}

		go reachOut(ad)
		addr = addr.Next()
	}
}

type CLI struct {
	Socket            string  `help:"The socket used for connecting to the client" default:"/tmp/thatch.sock" type:"string" short:"s"`
	ScanningPort      uint16  `help:"The port used for scanning for other users" default:"8000" type:"uint16" short:"d"`
	CommunicationPort uint16  `help:"The port used for communicating with other users" default:"8000" type:"uint16" short:"p"`
	SubnetMask        uint8   `help:"The subnet mask used for scanning for other users" default:"24" type:"uint8" short:"m"`
	Username          *string `help:"Username others will see" type:"string" short:"u"`
}

var cli CLI

func main() {

	var wg sync.WaitGroup

	kong.Parse(&cli)

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelInfo)

	// Find other users
	go discover()

	lis, err := net.Listen("unix", cli.Socket)
	if err != nil {
		slog.Error("Failed to listen: %v", err)
		return
	}

	defer os.Remove(cli.Socket)
	defer lis.Close()

	// Listen to client
	var opts []grpc.ServerOption
	intServer := grpc.NewServer(opts...)
	internal.RegisterInternalServiceServer(
		intServer,
		&internalServer{
			internal.UnimplementedInternalServiceServer{},
			&users, &chats,
		},
	)
	internal.RegisterControlServiceServer(
		intServer,
		&controlServer{wg: &wg},
	)

	go intServer.Serve(lis)
	defer intServer.Stop()
	slog.Info("Internal server started")

	// Listen to other users
	lis, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cli.CommunicationPort))
	if err != nil {
		slog.Error("Failed to listen: %v", err)
		return
	}

	extServer := grpc.NewServer(opts...)
	external.RegisterExternalServiceServer(
		extServer,
		&externalServer{
			external.UnimplementedExternalServiceServer{},
		},
	)
	go extServer.Serve(lis)
	defer extServer.Stop()
	slog.Info(fmt.Sprintf("External server started on port %d\n", cli.CommunicationPort))

	wg.Add(1)
	wg.Wait()

	slog.Info("Shutting down")

}
