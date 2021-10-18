package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"
)

const (
	defaultHost  = "localhost"
	defaultPort  = "3410"
	lenSuccessor = 3
)

var allCommands = make(map[string]func(string))

var ring = make(map[string]net.Conn)

type Node struct {
	finger      [161]string
	successor   [lenSuccessor]string
	predecessor string
}

type Feed struct {
	Messgae string
}

type handler func(*Feed)

type Server chan<- handler

type Nothing struct{}

func addCommand(cmd string, f func(string)) {
	allCommands[cmd] = f
}

func docommand(cmd string) {
	words := strings.Fields(cmd)

	if len(words) == 0 {
	} else {
		for i := 0; i < len(words); i++ {
			if f, exists := allCommands[strings.ToLower(words[i])]; exists {
				f(cmd)
			}
		}
	}
}

// InitiComand
func initcommands() {
	addCommand("help", cmdHelp)
	addCommand("quit", cmdQuit)
	addCommand("port", cmdPort)
	addCommand("create", cmdCreate)
	addCommand("join", cmdJoin)
}

func cmdHelp(s string) {
	fmt.Print("------Here is all the commands------\n")
	for key := range allCommands {
		fmt.Printf("		%s\n", key)
	}
	fmt.Print("------------------------------------\n")
}

func cmdPort(s string) {
	address := ":" + defaultPort
	words := strings.Fields(s)
	if len(words) > 1 {
		//fmt.Print((words[1]))
		address = ":" + words[1]
	}
	fmt.Print(address)
	server(address)
}

func cmdQuit(s string) {
	os.Exit(0)
}

func cmdCreate(s string) {

}

func cmdJoin(s string) {

}

func startActor() Server {
	ch := make(chan handler)
	state := new(Feed)
	go func() {
		for f := range ch {
			f(state)
		}
	}()
	return ch
}

func server(address string) {
	actor := startActor()
	rpc.Register(actor)
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Listen error :", err)
	}
	if err := http.Serve(l, nil); err != nil {
		log.Fatalf("http/Server: %v", err)
	}
}

func call(address string, method string, request interface{}, reply interface{}) error {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Fatalf("rpc.DialHTTP: %v", err)
	}
	defer client.Close()

	if err = client.Call(method, request, reply); err != nil {
		return err
	}

	return nil
}

func shell(address string) {

	log.Printf("starting interactive shell")
	log.Print("Commands are: help, port, create, join, quit, put, putrandom, get, delete")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		parts := strings.SplitN(line, " ", 2)
		if len(parts) > 1 {
			parts[1] = strings.TrimSpace(parts[1])
		}

		if len(parts) == 0 {
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Scanner error: %v", err)
	}
}

func main() {
	initcommands()

	fmt.Printf("Local address: %s\n", getLocalAddress())

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		//fmt.Printf("The line was: %s\n", line)
		docommand(line)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("Read command error")
	}
}

func getLocalAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
