package main

import (
	"bufio"
	"flag"
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

type Feed struct {
	Messgae string
}

type handler func(*Feed)

type Server chan<- handler

type Nothing struct{}

type Node struct {
	//finger      [161]string
	successor   [lenSuccessor]string
	predecessor string
}

func (s Server) Help(in *Nothing, out *[]string) error {
	finished := make(chan struct{})
	s <- func(f *Feed) {
		f.Messgae = "Commands are: help, port, create, join, quit, put, putrandom, get, delete"
		*out = append(*out, f.Messgae)
		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (s Server) Port(in string, out *Nothing) error {
	finished := make(chan struct{})
	s <- func(f *Feed) {
		newPort := defaultHost + ":" + in
		shell(newPort)
		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (s Server) create() error {
	finished := make(chan struct{})
	s <- func(f *Feed) {

		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (s Server) join() error {
	finished := make(chan struct{})
	s <- func(f *Feed) {

		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (s Server) Quit(in *Nothing, out *Nothing) error {
	finished := make(chan struct{})
	s <- func(f *Feed) {
		os.Exit(1)
		finished <- struct{}{}
	}
	<-finished
	return nil
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

func main() {
	var isServer bool
	var isClient bool
	var address string
	flag.BoolVar(&isServer, "server", false, "start as server")
	flag.BoolVar(&isClient, "client", false, "start as client")
	flag.Parse()

	if isServer && isClient {
		log.Fatalf("cannot be both client and server")
	}
	if !isClient && !isServer {
		printUsage()
	}

	switch flag.NArg() {
	case 0:
		if isClient {
			address = defaultHost + ":" + defaultPort
		} else {
			address = ":" + defaultPort
		}
	case 1:
		// user specified the address
		address = flag.Arg(0)
	default:
		printUsage()
	}

	if isClient {
		shell(address)
	} else {
		server(address)
	}
}

func printUsage() {
	log.Printf("Usage: %s [-server or -client] [address]", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
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

		switch parts[0] {
		case "help":
			var junk Nothing
			var message []string
			if err := call(address, "Server.Help", &junk, &message); err != nil {
				log.Fatalf("calling Server.Help: %v", err)
			}
			for _, elt := range message {
				log.Printf(elt)
			}

		case "port":
			var junk Nothing
			if err := call(address, "Server.Port", parts[1], &junk); err != nil {
				log.Fatalf("calling Server.Port: %v", err)
			}

		case "create":

		case "join":

		case "quit":
			var junkIn Nothing
			var junkOut Nothing
			if err := call(address, "Server.Quit", &junkIn, &junkOut); err != nil {
				log.Fatalf("calling Server.Quit: %v", err)
			}

		case "put":

		case "putrandom":

		case "get":

		case "delete":

		default:
			log.Printf("I onyly reconize: help, port, create, join, quit, put, putrandom, get, delete")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Scanner error: %v", err)
	}
}
