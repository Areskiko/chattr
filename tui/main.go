package main

import (
	"fmt"
	"net"
	"sync"

	"github.com/Areskiko/chattr/slib"
	"github.com/gen2brain/beeep"
	"github.com/rivo/tview"
)

var (
	active           int = 0
	chats                = make([]shared.Chat, 1)
	connectionsMutex sync.Mutex
)

var (
	app      *tview.Application
	chatList *tview.List
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
	mutex := sync.Mutex{}
	newChat := slib.NewChat{[]string{recipient}, conn, make([]slib.NewMessage, 8), &mutex}

	connectionsMutex.Lock()
	id := len(chats)
	chats = append(chats, newChat)
	connectionsMutex.Unlock()

	app.QueueUpdateDraw(func() {
		chatList.AddItem(recipient, fmt.Sprintf("Connected to %s", recipient), '0', func() {
			active = id
		})
	})

	for {
		_, err := conn.Read(buffer)
		if err != nil {
			return
		}

		content := string(buffer)

		newChat.lock.Lock()
		newChat.messages = append(newChat.messages, slib.NewMessage{recipient, content})
		newChat.lock.Unlock()
	}

}

func listenForConnections() {
	// Listen on port 8080
	listener, err := net.Listen("tcp", ":8080")
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

func main() {
	err := beeep.Notify("Title", "Message body", "assets/information.png")
	if err != nil {
		panic(err)
	}
	// Start the listener
	go listenForConnections()

	app = tview.NewApplication()
	chatList = tview.NewList()

	if err := app.SetRoot(chatList, true).SetFocus(chatList).Run(); err != nil {
		panic(err)
	}

}
