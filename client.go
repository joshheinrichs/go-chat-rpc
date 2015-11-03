package main

import (
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"
	"sync"
)

const (
	CMD_PREFIX = "/"
	CMD_CREATE = CMD_PREFIX + "create"
	CMD_LIST   = CMD_PREFIX + "list"
	CMD_JOIN   = CMD_PREFIX + "join"
	CMD_LEAVE  = CMD_PREFIX + "leave"
	CMD_HELP   = CMD_PREFIX + "help"
	CMD_NAME   = CMD_PREFIX + "name"
	CMD_QUIT   = CMD_PREFIX + "quit"

	MSG_HELP = "\nCommands:\n" +
		CMD_CREATE + " foo - creates a chat room named foo\n" +
		CMD_LIST + " - lists all chat rooms\n" +
		CMD_JOIN + " foo - joins a chat room named foo\n" +
		CMD_LEAVE + " - leaves the current chat room\n" +
		CMD_HELP + " - lists all commands\n" +
		CMD_NAME + " foo - changes your name to foo\n" +
		CMD_QUIT + " - quits the program\n\n"
)

var token string
var client *rpc.Client
var wg sync.WaitGroup

// Adds strings from stdin to the server.
func Input() {
	reader := bufio.NewReader(os.Stdin)
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		Parse(str)
	}
}

func Parse(str string) (err error) {
	switch {
	default:
		err = client.Call("Client.SendMessage", &Args{token, str}, nil)
	case strings.HasPrefix(str, CMD_CREATE):
		name := strings.TrimSuffix(strings.TrimPrefix(str, CMD_CREATE+" "), "\n")
		err = client.Call("Client.CreateChatRoom", &Args{token, name}, nil)
	case strings.HasPrefix(str, CMD_LIST):
		err = client.Call("Client.ListChatRooms", &token, nil)
	case strings.HasPrefix(str, CMD_JOIN):
		name := strings.TrimSuffix(strings.TrimPrefix(str, CMD_JOIN+" "), "\n")
		err = client.Call("Client.JoinChatRoom", &Args{token, name}, nil)
	case strings.HasPrefix(str, CMD_LEAVE):
		err = client.Call("Client.LeaveChatRoom", &token, nil)
	case strings.HasPrefix(str, CMD_NAME):
		name := strings.TrimSuffix(strings.TrimPrefix(str, CMD_NAME+" "), "\n")
		err = client.Call("Client.ChangeName", &Args{token, name}, nil)
	case strings.HasPrefix(str, CMD_HELP):
		fmt.Print(MSG_HELP)
	case strings.HasPrefix(str, CMD_QUIT):
		err = client.Call("Client.Quit", &token, nil)
	}
	return err
}

// Requests strings from the server and outputs them to stdout.
func Output() {
	for {
		var message string
		err := client.Call("Client.ReceiveMessage", &token, &message)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(message)
	}
}

func main() {
	wg.Add(1)

	var err error
	client, err = rpc.DialHTTP(CONN_TYPE, CONN_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	err = client.Call("Client.Connect", struct{}{}, &token)
	if err != nil {
		log.Fatal("connection error:", err)
	}
	fmt.Printf("token: %s\n", token)

	go Input()
	go Output()

	wg.Wait()
}
