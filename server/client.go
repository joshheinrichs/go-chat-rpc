package main

import (
	"errors"
	"sync"
)

var clients map[string]*Client = make(map[string]*Client)
var clientsMutex sync.RWMutex

func AddClient(client *Client) error {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	otherClient := clients[client.Token]
	if otherClient != nil {
		return errors.New(ERROR_TOKEN)
	}
	clients[client.Token] = client
	return nil
}

func GetClient(token string) (*Client, error) {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

	client := clients[token]
	if client == nil {
		return nil, errors.New(ERROR_NO_TOKEN)
	}
	return client, nil
}

func RemoveClient(token string) error {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	client := clients[token]
	if client == nil {
		return errors.New(ERROR_NO_TOKEN)
	}
	delete(clients, token)
	return nil
}

type Client struct {
	Token    string
	Name     string
	ChatRoom *ChatRoom
	Outgoing chan string
	Mutex    sync.RWMutex
}

func NewClient(token string) *Client {
	return &Client{
		Token:    token,
		Name:     "Anonymous",
		ChatRoom: nil,
		Outgoing: make(chan string),
	}
}
