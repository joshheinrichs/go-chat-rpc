package main

import (
	"../shared"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"time"
)

// randomString returns a random string with the specified length
func randomString(length int) (str string) {
	b := make([]byte, length)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
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
