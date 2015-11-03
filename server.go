package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net"
	"net/http"
	"net/rpc"
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

func (t *Client) Help(token *string, _ *struct{}) error {
	log.Println("Help")
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
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil)
}
