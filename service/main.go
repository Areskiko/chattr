package service

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/netip"

	external "github.com/areskiko/thatch/proto/inter"
	internal "github.com/areskiko/thatch/proto/intra"
	"google.golang.org/grpc"
)

var (
	username          string = "nordmann#0000"
	communicationPort uint16 = 8001
	scanningPort      uint16 = 8000
	subnetMask        uint8  = 24
	chats             map[string]*internal.Chat
	users             map[string]*internal.User
)

const INTERNAL_SOCKET = "/tmp/thatch.sock"

func reachOut(addr *net.TCPAddr) {
}

func discover() {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal("Could not create connection for discovery", err)
		return
	}

	defer conn.Close()

	local := conn.LocalAddr().(*net.UDPAddr)
	ip := local.IP
	cidr := fmt.Sprintf("%s/%d", ip, subnetMask)
	p, err := netip.ParsePrefix(cidr)
	if err != nil {
		log.Fatalf("Could not parse CIDR: %s", err)
	}

	p = p.Masked()
	log.Printf("Scanning %s", p.String())

	addr := p.Addr()
	for {
		if !p.Contains(addr) {
			break
		}

		ad, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", addr.String(), scanningPort))
		if err != nil {
			log.Printf("Error resolving %s: %s", addr.String(), err)
			return
		}

		go reachOut(ad)
		addr = addr.Next()
	}
}

func main() {
	username = *flag.String("u", "nordmann", "Username to display for others")
	communicationPort = uint16(*flag.Uint("p", 8001, "The port used to communicate with other users"))
	scanningPort = uint16(*flag.Uint("d", 8000, "The port others can use to discover you"))
	subnetMask = uint8(*flag.Uint("m", 24, "The subnet mask"))

	// Find other users
	go discover()

	lis, err := net.Listen("unix", INTERNAL_SOCKET)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Listen to client
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	internal.RegisterInternalServer(
		grpcServer,
		&internalServer{
			internal.UnimplementedInternalServer{},
			&users, &chats,
		},
	)
	go grpcServer.Serve(lis)

	// Listen to other users
	lis, err = net.Listen("tcp", fmt.Sprintf(":%d", communicationPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer = grpc.NewServer(opts...)
	external.RegisterExternalServer(
		grpcServer,
		&externalServer{
			external.UnimplementedExternalServer{},
		},
	)
	go grpcServer.Serve(lis)

	log.Println("Service started")
}
