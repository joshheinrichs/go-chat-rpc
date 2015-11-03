package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

const (
	MAX_CLIENTS = 10

	CLIENT_NAME = "Anonymous"

	ERROR_PREFIX = "Error: "
	ERROR_SEND   = ERROR_PREFIX + "You cannot send messages in the lobby.\n"
	ERROR_CREATE = ERROR_PREFIX + "A chat room with that name already exists.\n"
	ERROR_JOIN   = ERROR_PREFIX + "A chat room with that name does not exist.\n"
	ERROR_LEAVE  = ERROR_PREFIX + "You cannot leave the lobby.\n"

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

var users map[string]*User = make(map[string]*User)

type Client int

func (t *Client) Connect(args *struct{}, token *string) error {
	log.Println("Connect")
	*token = randomString(64)
	user := NewUser(*token)
	users[*token] = user

	return nil
}

func (t *Client) SendMessage(args *Args, _ *struct{}) error {
	log.Println("SendMessage")
	user := users[args.Token]
	user.Outgoing <- args.String
	log.Print(args.String)
	return nil
}

func (t *Client) CreateChatRoom(args *Args, _ *struct{}) error {
	log.Println("CreateChatRoom")
	return nil
}

func (t *Client) ListChatRooms(args *string, _ *struct{}) error {
	log.Println("ListChatRooms")
	return nil
}

func (t *Client) JoinChatRoom(args *Args, _ *struct{}) error {
	log.Println("JoinChatRoom")
	return nil
}

func (t *Client) LeaveChatRoom(token *string, _ *struct{}) error {
	log.Println("LeaveChatRoom")
	return nil
}

func (t *Client) ChangeName(args *Args, _ *struct{}) error {
	log.Println("ChangeName")
	return nil
}

func (t *Client) Quit(token *string, _ *struct{}) error {
	log.Println("Quit")
	return nil
}

func (t *Client) ReceiveMessage(token *string, message *string) error {
	log.Println("ReceiveMessage")
	user := users[*token]
	*message = <-user.Outgoing
	return nil
}

type User struct {
	Token    string
	Name     string
	Outgoing chan string
}

func NewUser(token string) *User {
	return &User{
		Token:    token,
		Name:     "Anonymous",
		Outgoing: make(chan string),
	}
}

// randomString returns a random string with the specified length
func randomString(length int) (str string) {
	b := make([]byte, length)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func main() {
	client := new(Client)
	rpc.Register(client)
	rpc.HandleHTTP()
	l, e := net.Listen(CONN_TYPE, CONN_PORT)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil)
}
