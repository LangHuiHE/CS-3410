package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

const listenAddress = ":3410"

var ch = make(chan bool)

type InputEvent struct {
	// player *Player
	Connection net.Conn
	Command    string
}

type OutputEvent struct {
	Connection net.Conn
	Command    string
}

func main() {
	// setup commands
	// setup world
	inputChannel := make(chan InputEvent)
	outputChannel := make(chan OutputEvent)
	go listenForConnections(inputChannel, outputChannel)
	mainLoop(inputChannel)
}

func mainLoop(inputChannel chan InputEvent) {
	for event := range inputChannel {
		//fmt.Fprintf(event.Connection,"event(on MUD):%s\n", event.Command)
		fmt.Printf("event(on MUD):%s\n", event.Command)
	}
}

func listenForConnections(inputChannel chan InputEvent, outputChannel chan OutputEvent) {
	listen, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatalf("listening: on %s: %v", listenAddress, err)
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatalf("accept: %v", err)
		}

		go handleConnection(conn, inputChannel, outputChannel)
	}
}

func handleConnection(conn net.Conn, inputChannel chan InputEvent, outputChannel chan OutputEvent) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		event := InputEvent{Connection: conn, Command: line}
		inputChannel <- event
		out_event := OutputEvent{Connection: conn, Command: line}
		outputChannel <- out_event
	}
	if err := scanner.Err(); err != nil {
		log.Printf("connection error: %v", err)
	} else {
		log.Printf("connection closed normally\n")
	}

	for out_event := range outputChannel {
		fmt.Fprintf(out_event.Connection, "event(on Network):%s\n", out_event.Command)
	}
}
