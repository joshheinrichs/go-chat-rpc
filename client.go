package main

import (
	"fmt"
	"log"
	"net/rpc"
)

// Requests strings from the server and outputs them to stdout.
func Read() {

}

// Adds strings from stdin to the server.
func Write() {

}

func main() {
	client, err := rpc.DialHTTP("tcp", ":1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	// Synchronous call
	args := &Args{7, 8}
	var reply int
	err = client.Call("Arith.Multiply", args, &reply)
	if err != nil {
		log.Fatal("arith error:", err)
	}
	fmt.Printf("Arith: %d*%d=%d\n", args.A, args.B, reply)
}
