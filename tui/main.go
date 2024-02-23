package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	pb "github.com/areskiko/thatch/proto"
)

const TRUNCATION_LENGTH = 10

/* Creating new chats */

type Id string

type Available struct {
	username string
	bio      string
	id       Id
}

func (a Available) FilterValue() string {
	return a.username
}

func (a Available) Title() string {
	return a.username
}

func (a Available) Description() string {
	return a.bio
}

/* Interacting with chats */

type Message struct {
	sender  Id
	content string
}

type Chat struct {
	username string
	messages []Message
	id       Id
}

func (a Chat) FilterValue() string {
	return a.username
}

func (a Chat) Title() string {
	return a.username
}

func (a Chat) Description() string {
	last := a.messages[len(a.messages)-1].content
	var truncated string

	if len(last) > TRUNCATION_LENGTH {
		var truncated_builder strings.Builder
		for i := 0; i < TRUNCATION_LENGTH-3; i++ {
			truncated_builder.WriteByte(last[i])
		}
		truncated_builder.WriteString("...")
		truncated = truncated_builder.String()
	} else {
		truncated = last
	}

	return truncated
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
}

func New() *Model {
	return &Model{}
}

func (m *Model) initList(width int, height int) {
	m.available_users = list.New([]list.Item{}, list.NewDefaultDelegate(), width, height)
	m.available_users.Title = "Available users"
	m.available_users.SetItems([]list.Item{
		Available{username: "User1", bio: "Cool bio 1", id: "1"},
		Available{username: "User2", bio: "Cool bio 2", id: "2"},
		Available{username: "User3", bio: "Cool bio 3", id: "3"},
		Available{username: "User4", bio: "Cool bio 4", id: "4"},
		Available{username: "User5", bio: "Cool bio 5", id: "5"},
	})

	m.existing_chats = list.New([]list.Item{}, list.NewDefaultDelegate(), width, height)
	m.existing_chats.Title = "Active Chats"
	m.existing_chats.SetItems([]list.Item{
		Chat{
			username: "User1",
			messages: []Message{{content: "Cool story 1", sender: "1"}, {content: "This is a message", sender: "1"}, {content: "This is my reply", sender: "0"}, {content: "A follow up", sender: "1"}},
			id:       "1"},
		Chat{
			username: "User2",
			messages: []Message{{content: "Cool story 2", sender: "2"}, {content: "This is a message", sender: "2"}},
			id:       "2"},
	})

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
	// TODO: Connect to server
	// TODO: Spawn server if doesn't exist
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var commands []tea.Cmd = make([]tea.Cmd, 2)
	var command tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.initList(msg.Width, msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "l":
			m.state = available_users
		case "c":
			m.state = existing_chats
		case "enter", " ":
			switch m.state {
			case available_users:
				// TODO: Ask server to do just that
				fmt.Printf("Creating chat with user")
			case existing_chats:
				// TODO: Refresh messages in chat
				m.state = chat
				m.chat_index = m.existing_chats.Index()
				items := m.existing_chats.Items()
				chat := items[m.chat_index].(Chat)
				messages := make([]string, len(chat.messages))
				for _, v := range chat.messages {
					messages = append(messages, fmt.Sprintf("%s: %s", v.sender, v.content))
				}

				m.chat_viewport.SetContent(strings.Join(messages, "\n"))
				m.chat_viewport.GotoBottom()
				fmt.Printf("Selected chat %d", m.chat_index)
				//commands = append(commands, m.chat_textarea.Cursor.BlinkCmd())
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
	title.WriteString(m.existing_chats.Items()[m.chat_index].(Chat).username)

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
	m := New()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
