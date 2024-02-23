package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"net/netip"
	"strings"
	"sync"
	"time"

	pb "github.com/areskiko/thatch/proto"
)

var (
	username      string       = "nordmann#0000"
	mainPort      uint16       = 8001
	discoveryPort uint16       = 8000
	subnetMask    uint8        = 24
	chats         []pb.Chat = make([]pb.Chat, 0)
	chatsMutex    sync.Mutex
	users         []pb.User = make([]pb.User, 0)
	usersMutex    sync.Mutex
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Handle incoming data here
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		return
	}

	recipient := string(buffer)
	newChat := slib.NewChat([]string{recipient}, conn)

	chatsMutex.Lock()
	chats = append(chats, newChat)
	chatsMutex.Unlock()

	for {
		_, err := conn.Read(buffer)
		if err != nil {
			return
		}

		content := string(buffer)

		newChat.Push(recipient, content)
	}

}

func makeDiscoverable() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", discoveryPort))
	if err != nil {
		return
	}

	defer listener.Close()

	for {
		// Wait for a connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		defer conn.Close()

		msg := make([]byte, 2)
		binary.LittleEndian.PutUint16(msg, mainPort)

		conn.Write(msg)
		conn.Write([]byte(username))
	}
}

func listenForConnections() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", mainPort))
	if err != nil {
		return
	}
	defer listener.Close()

	for {
		// Wait for a connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func reachOut(address net.Addr) {
	msg := make([]byte, 2)
	conn, err := net.Dial(address.Network(), address.String())

	if err != nil {
		return
	}

	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Read(msg)
	if err != nil {
		return
	}

	port := binary.LittleEndian.Uint16(msg)

	_, err = conn.Read(msg)
	if err != nil {
		return
	}

	user := string(msg)
	sport := fmt.Sprintf("%s:%d", strings.Split(address.String(), ":")[0], port)
	addr, err := net.ResolveTCPAddr("tcp", sport)
	if err != nil {
		return
	}

	usersMutex.Lock()
	users = append(users, slib.NewUser(user, addr))
	usersMutex.Unlock()
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

		ad, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", addr.String(), discoveryPort))
		if err != nil {
			log.Printf("Error resolving %s: %s", addr.String(), err)
			return
		}

		go reachOut(ad)
		addr = addr.Next()
	}
}

func controll() {
	listener, err := net.Listen("unix", "/tmp/chattr")
	if err != nil {
		log.Fatalf("Failed to create local listener: %s", err)
	}
	defer listener.Close()

	buffer := make([]byte, 1024)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection", err)
			continue
		}

		defer conn.Close()

		for {

			_, err = conn.Read(buffer)
			if err != nil {
				fmt.Println("Error reading command", err)
				break
			}

			// TODO: handle commands
			content := string(buffer)
			conn.Write([]byte(fmt.Sprintf("Received: %s", content)))

		}

	}
}

func main() {

	username = *flag.String("u", "nordmann", "Username to display for others")
	mainPort = uint16(*flag.Uint("p", 8001, "The port used to communicate with other users"))
	discoveryPort = uint16(*flag.Uint("d", 8000, "The port others can use to discover you"))
	subnetMask = uint8(*flag.Uint("m", 24, "The subnet mask"))

	log.Println("Service started")
	go listenForConnections()
	go makeDiscoverable()
	go discover()

	controll()
}
