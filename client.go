package main

import (
	"./shared"
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
	MSG_DISCONNECT = "Disconnected from the server.\n"
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
			wg.Done()
			break
		}
		Parse(str)
	}
}

func Parse(str string) (err error) {
	switch {
	default:
		err = client.Call("Receiver.SendMessage", &shared.Args{token, str}, nil)
	case strings.HasPrefix(str, CMD_CREATE):
		name := strings.TrimSuffix(strings.TrimPrefix(str, CMD_CREATE+" "), "\n")
		err = client.Call("Receiver.CreateChatRoom", &shared.Args{token, name}, nil)
	case strings.HasPrefix(str, CMD_LIST):
		err = client.Call("Receiver.ListChatRooms", &token, nil)
	case strings.HasPrefix(str, CMD_JOIN):
		name := strings.TrimSuffix(strings.TrimPrefix(str, CMD_JOIN+" "), "\n")
		err = client.Call("Receiver.JoinChatRoom", &shared.Args{token, name}, nil)
	case strings.HasPrefix(str, CMD_LEAVE):
		err = client.Call("Receiver.LeaveChatRoom", &token, nil)
	case strings.HasPrefix(str, CMD_NAME):
		name := strings.TrimSuffix(strings.TrimPrefix(str, CMD_NAME+" "), "\n")
		err = client.Call("Receiver.ChangeName", &shared.Args{token, name}, nil)
	case strings.HasPrefix(str, CMD_HELP):
		fmt.Print(MSG_HELP)
	case strings.HasPrefix(str, CMD_QUIT):
		err = client.Call("Receiver.Quit", &token, nil)
		wg.Done()
	}
	return err
}

// Requests strings from the server and outputs them to stdout.
func Output() {
	for {
		var message string
		err := client.Call("Receiver.ReceiveMessage", &token, &message)
		if err != nil {
			wg.Done()
			break
		}
		fmt.Print(message)
	}
}

func main() {
	wg.Add(1)

	var err error
	client, err = rpc.DialHTTP(shared.CONN_TYPE, shared.CONN_PORT)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Call("Receiver.Connect", &struct{}{}, &token)
	if err != nil {
		log.Fatal(err)
	}

	go Input()
	go Output()

	wg.Wait()
	fmt.Print(MSG_DISCONNECT)
}
