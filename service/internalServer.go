package main

import (
	"context"
	"fmt"

	internal "github.com/areskiko/thatch/proto/intra"
)


type internalServer struct {
	internal.UnimplementedInternalServiceServer
	users *map[string]*internal.User
	chats *map[string]*internal.Chat
}

func (s *internalServer) GetChat(context context.Context, request *internal.GetChatRequest) (*internal.GetChatResponse, error) {
	id := request.GetChatId()
	chat := (*s.chats)[id]

	if chat == nil {
		return nil, fmt.Errorf("Chat with id %s not found", id)
	}

	return &internal.GetChatResponse{Chat: chat}, nil
}

func (s *internalServer) GetChats(context.Context, *internal.GetChatsRequest) (*internal.GetChatsResponse, error) {
	ids := &internal.GetChatsResponse{}
	for id := range *s.chats {
		ids.ChatIds = append(ids.ChatIds, id)
	}
	return ids, nil
}

// GetUsers implements proto.InternalServerServer.
func (s *internalServer) GetUsers(context.Context, *internal.GetUsersRequest) (*internal.GetUsersResponse, error) {
	usrs := &internal.GetUsersResponse{}
	for _, user := range *s.users {
		usrs.Users = append(usrs.Users, user)
	}

	return usrs, nil
}

// SendMessage implements proto.InternalServerServer.
func (s *internalServer) SendMessage(context.Context, *internal.SendMessageRequest) (*internal.SendMessageResponse, error) {
	println("unimplemented")
	return nil, nil
}

// StartChat implements proto.InternalServerServer.
func (s *internalServer) StartChat(context.Context, *internal.StartChatRequest) (*internal.StartChatResponse, error) {
	println("unimplemented")
	return nil, nil
}
