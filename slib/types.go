package slib

import (
	"net"
	"sync"
)

type Message struct {
	sender  string
	content string
}

func NewMessage(sender string, content string) *Message {
	return &Message{sender: sender, content: content}
}

type Chat struct {
	members    []string
	connection net.Conn
	messages   []Message
	lock       sync.Mutex
}

func NewChat(recipient []string, connection net.Conn) *Chat {
	chat := Chat{members: recipient, connection: connection, messages: make([]Message, 8), lock: sync.Mutex{}}
	return &chat
}

func (c *Chat) PushMessage(message Message) {
	c.lock.Lock()
	c.messages = append(c.messages, message)
	c.lock.Unlock()
}

func (c *Chat) Push(sender string, content string) {
	c.PushMessage(*NewMessage(sender, content))
}
