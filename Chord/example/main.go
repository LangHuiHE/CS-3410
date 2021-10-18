package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"strings"
)

const (
	defaultHost = "localhost"
	defaultPort = "3410"
)

type Feed struct {
	Messages []string
}

type handler func(*Feed)
type Server chan<- handler

type Nothing struct{}

func (s Server) Post(msg string, reply *Nothing) error {
	finished := make(chan struct{})
	s <- func(f *Feed) {
		f.Messages = append(f.Messages, msg)
		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (s Server) Get(count int, reply *[]string) error {
	finished := make(chan struct{})
	s <- func(f *Feed) {
		if len(f.Messages) < count {
			count = len(f.Messages)
		}
		*reply = make([]string, count)
		copy(*reply, f.Messages[len(f.Messages)-count:])
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
		log.Fatalf("cannot be both a client and a server")
	}
	if !isServer && !isClient {
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

	l, e := net.Listen("tcp", address)
	if e != nil {
		log.Fatal("Listen error:", e)
	}

	if err := http.Serve(l, nil); err != nil {
		log.Fatalf("http/Serve: %v", err)
	}
}

func client(address string) {
	var junk Nothing
	if err := call(address, "Server.Post", "Hellow, again", &junk); err != nil {
		log.Fatalf("client.Call: %v", err)
	}
	if err := call(address, "Server.Post", "I ate cereal for breakfast", &junk); err != nil {
		log.Fatalf("client.Call: %v", err)
	}

	var lst []string
	if err := call(address, "Server.Get", 5, &lst); err != nil {
		log.Fatalf("client.Call: %v", err)
	}
	for _, elt := range lst {
		log.Println(elt)
	}
}

func call(serverAddress string, method string, request interface{}, responce interface{}) error {
	client, err := rpc.DialHTTP("tcp", serverAddress)
	if err != nil {
		log.Fatalf("rpc.DialHTTP: %v", err)
		return err
	}
	defer client.Close()

	if err = client.Call(method, request, responce); err != nil {
		return err
	}

	return nil
}

func shell(address string) {
	log.Printf("Starting interactive shell")
	log.Printf("Commands are: get, post")

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
		case "get":
			n := 10
			if len(parts) == 2 {
				var err error
				if n, err = strconv.Atoi(parts[1]); err != nil {
					log.Fatalf("parsing number of messages: %v", err)
				}
			}

			var messages []string
			if err := call(address, "Server.Get", n, &messages); err != nil {
				log.Fatalf("calling Server.Get: %v", err)
			}
			for _, elt := range messages {
				log.Println(elt)
			}

		case "post":
			if len(parts) != 2 {
				log.Printf("you must specify a message to post")
				continue
			}

			var junk Nothing
			if err := call(address, "Server.Post", parts[1], &junk); err != nil {
				log.Fatalf("calling Server.Post: %v", err)
			}

		default:
			log.Printf("I only recognize \"get\" and \"post\"")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("scanner error: %v", err)

	}
}
