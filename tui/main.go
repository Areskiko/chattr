package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	pb "github.com/areskiko/thatch/proto/intra"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"google.golang.org/grpc"
)

const TRUNCATION_LENGTH = 10
const SOCKET = "/tmp/thatch.sock"

/* Creating new chats */

type Id string

type Available struct {
	user *pb.User
}

func (a Available) FilterValue() string {
	return a.user.GetName()
}

func (a Available) Title() string {
	return a.user.GetName()
}

func (a Available) Description() string {
	return ""
}

/* Interacting with chats */

type Message struct {
	sender  Id
	content string
}

type Chat struct {
	id string
}

func (a Chat) FilterValue() string {
	return a.id
}

func (a Chat) Title() string {
	return a.id
}

func (a Chat) Description() string {
	return ""
}

/* App */

type state int

const (
	available_users state = iota
	existing_chats
	chat
)

var (
	View       state
	ActiveChat Id
)

type Model struct {
	state           state
	chat_index      int
	chat_viewport   viewport.Model
	chat_textarea   textarea.Model
	available_users list.Model
	existing_chats  list.Model
	err             error
	client          pb.InternalClient
}

type ConnectedMessage struct{}

func New() *Model {
	return &Model{}
}

func (m *Model) initList(width int, height int) {
	m.available_users = list.New([]list.Item{}, list.NewDefaultDelegate(), width, height)
	m.available_users.Title = "Available Users"

	m.existing_chats = list.New([]list.Item{}, list.NewDefaultDelegate(), width, height)
	m.existing_chats.Title = "Active Chats"

	ta := textarea.New()
	ta.Placeholder = "Send a message..."

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	m.chat_textarea = ta

	m.chat_viewport = viewport.New(30, 5)
}

func (m Model) Init() tea.Cmd {
	// TODO: Spawn server if doesn't exist
	connect := func() tea.Msg {

		var opts []grpc.DialOption

		conn, err := grpc.Dial(SOCKET, opts...)
		if err != nil {
			m.err = err
		}

		client := pb.NewInternalClient(conn)
		m.client = client
		return ConnectedMessage{}
	}

	return tea.Batch(connect)
}

func fetchAvailableUsers(m *Model) func() tea.Msg {
	return func() tea.Msg {
		users, err := m.client.GetUsers(context.Background(), &pb.Empty{})
		if err != nil {
			m.err = err
		}

		items := make([]list.Item, len(users.GetUsers()))
		for _, user := range users.Users {
			available := Available{
				user: user,
			}
			items = append(items, available)
		}
		m.available_users.SetItems(items)

		return nil
	}
}

func fetchExistingChats(m *Model) func() tea.Msg {
	return func() tea.Msg {
		chats, err := m.client.GetChats(context.Background(), &pb.Empty{})
		if err != nil {
			m.err = err
		}

		items := make([]list.Item, len(chats.GetChatIds()))
		for _, chatId := range chats.GetChatIds() {
			chat := Chat{
				id: chatId,
			}
			items = append(items, chat)
		}
		m.existing_chats.SetItems(items)

		return nil
	}
}

func loadChat(m *Model) func() tea.Msg {
	return func() tea.Msg {
		result, err := m.client.GetChat(context.Background(), &pb.GetChatRequest{ChatId: m.existing_chats.Items()[m.chat_index].(Chat).id})
		if err != nil {
			m.err = err
		}

		messages := make([]string, len(result.GetMessages()))
		for _, message := range result.GetMessages() {
			messages = append(messages, fmt.Sprintf("%s: %s", message.GetSender(), message.GetText()))
		}
		m.chat_viewport.SetContent(strings.Join(messages, "\n"))
		m.chat_viewport.GotoBottom()

		return nil
	}
}

func createChat(m *Model) func() tea.Msg {
	return func() tea.Msg {
		chat, err := m.client.StartChat(context.Background(), &pb.User{Id: m.available_users.Items()[m.available_users.Index()].(Available).user.GetId()})
		if err != nil {
			m.err = err
		}

		fetchExistingChats(m)()

		m.state = existing_chats
		for i, item := range m.existing_chats.Items() {
			if item.(Chat).id == chat.GetId() {
				m.chat_index = i
				break
			}
		}

		loadChat(m)()

		return nil
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var commands []tea.Cmd = make([]tea.Cmd, 2)
	var command tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.initList(msg.Width, msg.Height)

	case ConnectedMessage:
		command = fetchAvailableUsers(&m)
		commands = append(commands, command)

	case tea.KeyMsg:
		switch msg.String() {
		case "l":
			m.state = available_users

		case "c":
			m.state = existing_chats

		case "enter", " ":
			switch m.state {
			case available_users:
				command = createChat(&m)
				commands = append(commands, command)

			case existing_chats:
				m.state = chat
				m.chat_index = m.existing_chats.Index()
				command = loadChat(&m)
				commands = append(commands, command)
			}
		}
	}

	switch m.state {
	case available_users:
		m.available_users, command = m.available_users.Update(msg)
		commands = append(commands, command)
	case existing_chats:
		m.existing_chats, command = m.existing_chats.Update(msg)
		commands = append(commands, command)
	case chat:
		m.chat_textarea, command = m.chat_textarea.Update(msg)
		commands = append(commands, command)
		m.chat_viewport, command = m.chat_viewport.Update(msg)
		commands = append(commands, command)

	}

	// TODO: Refresh state from server

	return m, tea.Batch(commands...)
}

func (m Model) RenderChat() string {
	var output, title strings.Builder

	title.WriteString("Chat with ")
	title.WriteString(m.existing_chats.Items()[m.chat_index].(Chat).id)

	output.WriteString(lipgloss.NewStyle().Background(lipgloss.Color("#663399")).Render(title.String()))

	output.WriteString("\n\n")
	output.WriteString(m.chat_viewport.View())
	output.WriteString("\n\n")
	output.WriteString(lipgloss.PlaceVertical(35, lipgloss.Bottom, m.chat_textarea.View()))

	return lipgloss.NewStyle().MarginLeft(1).Render(output.String())
}

func (m Model) View() string {
	var output strings.Builder

	switch m.state {
	case available_users:
		output.WriteString(m.available_users.View())
	case existing_chats:
		output.WriteString(m.existing_chats.View())
	case chat:
		output.WriteString(m.RenderChat())
	default:
		output.WriteString(m.available_users.View())
	}
	return output.String()
}

func main() {
	var a pb.Empty

	fmt.Printf("Hello, world: %v", &a)

	m := New()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
