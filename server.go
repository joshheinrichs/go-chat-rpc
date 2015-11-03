package main

import (
	"./shared"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

const (
	MAX_CLIENTS = 10

	CLIENT_NAME = "Anonymous"

	ERROR_PREFIX   = "Error: "
	ERROR_SEND     = ERROR_PREFIX + "You cannot send messages in the lobby.\n"
	ERROR_CREATE   = ERROR_PREFIX + "A chat room with that name already exists.\n"
	ERROR_JOIN     = ERROR_PREFIX + "A chat room with that name does not exist.\n"
	ERROR_LEAVE    = ERROR_PREFIX + "You cannot leave the lobby.\n"
	ERROR_TOKEN    = ERROR_PREFIX + "A client with that token already exists.\n"
	ERROR_NO_TOKEN = ERROR_PREFIX + "No client exists with that token.\n"

	NOTICE_PREFIX          = "Notice: "
	NOTICE_ROOM_JOIN       = NOTICE_PREFIX + "\"%s\" joined the chat room.\n"
	NOTICE_ROOM_LEAVE      = NOTICE_PREFIX + "\"%s\" left the chat room.\n"
	NOTICE_ROOM_NAME       = NOTICE_PREFIX + "\"%s\" changed their name to \"%s\".\n"
	NOTICE_ROOM_DELETE     = NOTICE_PREFIX + "Chat room is inactive and being deleted.\n"
	NOTICE_PERSONAL_CREATE = NOTICE_PREFIX + "Created chat room \"%s\".\n"
	NOTICE_PERSONAL_NAME   = NOTICE_PREFIX + "Changed name to \"\".\n"

	MSG_CONNECT = "Welcome to the server! Type \"/help\" to get a list of commands.\n"
	MSG_FULL    = "Server is full. Please try reconnecting later."

	EXPIRY_TIME time.Duration = 7 * 24 * time.Hour
)

var clients map[string]*Client = make(map[string]*Client)
var clientsMutex sync.RWMutex

var chatRooms map[string]*ChatRoom = make(map[string]*ChatRoom)
var chatRoomsMutex sync.RWMutex

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

type Receiver int

func (r *Receiver) Connect(args *struct{}, token *string) error {
	log.Println("Connect")
	*token = randomString(64)
	client := NewClient(*token)
	err := AddClient(client)
	if err != nil {
		log.Println(err)
		return err
	}
	go func() { client.Outgoing <- MSG_CONNECT }()
	return nil
}

func (r *Receiver) SendMessage(args *shared.Args, _ *struct{}) error {
	log.Println("SendMessage")
	client, err := GetClient(args.Token)
	if err != nil {
		log.Println(err)
		return err
	}
	client.Mutex.RLock()
	defer client.Mutex.RUnlock()
	if client.ChatRoom == nil {
		client.Outgoing <- ERROR_SEND
		return nil
	}
	client.ChatRoom.Incoming <- fmt.Sprintf("%s - %s: %s", time.Now().Format(time.Kitchen), client.Name, args.String)
	return nil
}

func (r *Receiver) CreateChatRoom(args *shared.Args, _ *struct{}) error {
	log.Println("CreateChatRoom")
	client, err := GetClient(args.Token)
	if err != nil {
		log.Println(err)
		return err
	}
	chatRoom := NewChatRoom(args.String)
	err = AddChatRoom(chatRoom)
	if err != nil {
		client.Outgoing <- err.Error()
		log.Println(err)
		return err
	}
	client.Outgoing <- fmt.Sprintf(NOTICE_PERSONAL_CREATE, chatRoom.Name)
	return nil
}

func (r *Receiver) ListChatRooms(token *string, _ *struct{}) error {
	log.Println("ListChatRooms")
	client, err := GetClient(*token)
	if err != nil {
		log.Println(err)
		return err
	}
	chatRoomNames := GetChatRoomNames()
	chatList := "\nChatRooms:\n"
	for _, chatRoomName := range chatRoomNames {
		chatList += chatRoomName + "\n"
	}
	chatList += "\n"
	client.Outgoing <- chatList
	return nil
}

func (r *Receiver) JoinChatRoom(args *shared.Args, _ *struct{}) error {
	log.Println("JoinChatRoom")
	client, err := GetClient(args.Token)
	if err != nil {
		log.Println(err)
		return err
	}
	chatRoom, err := GetChatRoom(args.String)
	if err != nil {
		client.Outgoing <- err.Error()
		log.Println(err)
		return err
	}
	client.Mutex.RLock()
	oldChatRoom := client.ChatRoom
	client.Mutex.RUnlock()
	if oldChatRoom != nil {
		oldChatRoom.Leave <- client
	}

	chatRoom.Join <- client
	return nil
}

func (r *Receiver) LeaveChatRoom(token *string, _ *struct{}) error {
	log.Println("LeaveChatRoom")
	client, err := GetClient(*token)
	if err != nil {
		log.Println(err)
		return err
	}
	client.Mutex.RLock()
	defer client.Mutex.RUnlock()
	client.ChatRoom.Leave <- client
	return nil
}

func (r *Receiver) ChangeName(args *shared.Args, _ *struct{}) error {
	log.Println("ChangeName")
	client, err := GetClient(args.Token)
	if err != nil {
		return err
	}
	client.Mutex.Lock()
	defer client.Mutex.Unlock()
	client.Name = args.String
	return nil
}

func (r *Receiver) Quit(token *string, _ *struct{}) error {
	log.Println("Quit")
	err := RemoveClient(*token)
	if err != nil {
		return err
	}
	return nil
}

func (r *Receiver) ReceiveMessage(token *string, message *string) error {
	log.Println("ReceiveMessage")
	client, err := GetClient(*token)
	if err != nil {
		return err
	}
	*message = <-client.Outgoing
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

// randomString returns a random string with the specified length
func randomString(length int) (str string) {
	b := make([]byte, length)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func main() {
	rcvr := new(Receiver)
	rpc.Register(rcvr)
	rpc.HandleHTTP()
	l, e := net.Listen(shared.CONN_TYPE, shared.CONN_PORT)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil)
}
