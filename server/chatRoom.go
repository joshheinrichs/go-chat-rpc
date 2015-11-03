package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

var chatRooms map[string]*ChatRoom = make(map[string]*ChatRoom)
var chatRoomsMutex sync.RWMutex

func AddChatRoom(chatRoom *ChatRoom) error {
	chatRoomsMutex.Lock()
	defer chatRoomsMutex.Unlock()

	otherChatRoom := chatRooms[chatRoom.Name]
	if otherChatRoom != nil {
		return errors.New(ERROR_CREATE)
	}
	chatRooms[chatRoom.Name] = chatRoom
	return nil
}

func GetChatRoom(name string) (*ChatRoom, error) {
	chatRoomsMutex.RLock()
	defer chatRoomsMutex.RUnlock()

	chatRoom := chatRooms[name]
	if chatRoom == nil {
		return nil, errors.New(ERROR_JOIN)
	}
	return chatRoom, nil
}

func GetChatRoomNames() []string {
	chatRoomsMutex.RLock()
	defer chatRoomsMutex.RUnlock()

	keys := make([]string, 0, len(chatRooms))
	for k := range chatRooms {
		keys = append(keys, k)
	}
	return keys
}

func RemoveChatRoom(name string) error {
	chatRoomsMutex.Lock()
	defer chatRoomsMutex.Unlock()

	chatRoom := chatRooms[name]
	if chatRoom == nil {
		return errors.New(ERROR_JOIN)
	}
	delete(chatRooms, name)
	return nil
}

type ChatRoom struct {
	Name     string
	Clients  []*Client
	Messages []string
	Join     chan *Client
	Leave    chan *Client
	Incoming chan string
	Expire   chan bool
	Expiry   time.Time
}

func NewChatRoom(name string) *ChatRoom {
	chatRoom := &ChatRoom{
		Name:     name,
		Clients:  make([]*Client, 0),
		Messages: make([]string, 0),
		Join:     make(chan *Client),
		Leave:    make(chan *Client),
		Incoming: make(chan string),
		Expire:   make(chan bool),
		Expiry:   time.Now().Add(EXPIRY_TIME),
	}
	chatRoom.Listen()
	chatRoom.TryDelete()
	return chatRoom
}

func (chatRoom *ChatRoom) Listen() {
	go func() {
		for {
			select {
			case client := <-chatRoom.Join:
				chatRoom.AddClient(client)
			case client := <-chatRoom.Leave:
				chatRoom.RemoveClient(client)
			case message := <-chatRoom.Incoming:
				chatRoom.Broadcast(message)
			case _ = <-chatRoom.Expire:
				chatRoom.TryDelete()
			}
		}
	}()
}

func (chatRoom *ChatRoom) AddClient(client *Client) {
	log.Println("AddClient")
	client.Mutex.Lock()
	defer client.Mutex.Unlock()

	chatRoom.Broadcast(fmt.Sprintf(NOTICE_ROOM_JOIN, client.Name))
	for _, message := range chatRoom.Messages {
		client.Outgoing <- message
	}
	chatRoom.Clients = append(chatRoom.Clients, client)
	client.ChatRoom = chatRoom
}

func (chatRoom *ChatRoom) RemoveClient(client *Client) {
	log.Println("RemoveClient")
	client.Mutex.RLock()
	chatRoom.Broadcast(fmt.Sprintf(NOTICE_ROOM_LEAVE, client.Name))
	client.Mutex.RUnlock()
	for i, otherClient := range chatRoom.Clients {
		if client == otherClient {
			chatRoom.Clients = append(chatRoom.Clients[:i], chatRoom.Clients[i+1:]...)
			break
		}
	}
	client.Mutex.Lock()
	defer client.Mutex.Unlock()
	client.ChatRoom = nil
}

func (chatRoom *ChatRoom) Broadcast(message string) {
	log.Println("Broadcast")
	chatRoom.Expiry = time.Now().Add(EXPIRY_TIME)
	log.Println(message)
	chatRoom.Messages = append(chatRoom.Messages, message)
	for _, client := range chatRoom.Clients {
		client.Outgoing <- message
	}
}

func (chatRoom *ChatRoom) TryDelete() {
	log.Println("TryDelete")
	if chatRoom.Expiry.After(time.Now()) {
		go func() {
			time.Sleep(chatRoom.Expiry.Sub(time.Now()))
			chatRoom.Expire <- true
		}()
	} else {
		chatRoom.Broadcast(NOTICE_ROOM_DELETE)
		for _, client := range chatRoom.Clients {
			client.Mutex.Lock()
			client.ChatRoom = nil
			client.Mutex.Unlock()
		}
		RemoveChatRoom(chatRoom.Name)
		//TODO: Clear out channels
	}
}
