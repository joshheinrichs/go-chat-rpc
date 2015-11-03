package main

import (
	"../shared"
	"log"
	"net"
	"net/http"
	"net/rpc"
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
