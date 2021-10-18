package main

import (
	"log"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
)

type FindAnswer struct {
	OK        bool
	Successor string
}

type FindKey struct {
	ID    *big.Int
	Start string
}

func (n *Node) ClosePrecedingNode(id *big.Int) string {
	// fmt.Printf("len=%d cap=%d %v\n", len(n.finger), cap(n.finger), n.finger)
	for i := sizeOffinger - 1; i >= 1; i-- {
		if n.finger[i] != "" {
			if between(n.id, hashString(n.finger[i]), id, false) {
				return n.finger[i]
			}
		}
	}
	return n.address
}

func (n *Node) Find(id *big.Int, start string) string {
	found, nextNode := false, start
	i := 0

	for !found && i < maxSteps {

		var findSucReply FindAnswer
		if err := call(nextNode, "Server.FindSuccessor", id, &findSucReply); err != nil {
			log.Fatalf("calling Server.Find.FindSuccessor: %v\n", err)
		}
		found = findSucReply.OK
		nextNode = findSucReply.Successor
		i++
		if found {
			break
		}
	}
	if found {
		return nextNode
	} else {
		return "ERROR"
	}
}

// called periodically. refreshes finger table entries.
// next stores the index of the next finger to fix.
func (n *Node) FixFingers() {
	// fmt.Printf("len=%d cap=%d %v\n", len(n.finger), cap(n.finger), n.finger)
	nextSucIndex := 0
	for i := 1; i < sizeOffinger; i++ {
		if nextSucIndex < len(n.successor) {
			nextJump := jump(n.address, i)
			if nextJump.Cmp(hashString(n.successor[nextSucIndex])) == 1 {
				var reply FindAnswer
				if err := call(n.address, "Server.FindSuccessor", jump(n.address, i), &reply); err != nil {
					log.Fatalf("calling Server.FindSuccessor: %v\n", err)
				}
				n.finger[i] = reply.Successor
				nextSucIndex++
			} else {
				n.finger[i] = n.successor[nextSucIndex]
			}
		}
	}

}

// called periodically. checks whether predecessor has failed.
func (n *Node) CheckPredecessor() {
	var in, out Nothing
	if err := call(n.predecesor, "Server.Ping", &in, &out); err != nil {
		n.predecesor = ""
	}
}

func (s Server) Ping(in *Nothing, out *Nothing) error {
	finished := make(chan struct{})

	s <- func(n *Node) {
		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (s Server) GetAll(address string, out *map[string]string) error {
	finished := make(chan struct{})
	contain := make(map[string]string)

	s <- func(n *Node) {
		for key, value := range n.data {
			if between(hashString(address), hashString(key), n.id, false) {
				contain[key] = value
				delete(n.data, key)
			}
		}
		*out = contain
		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (s Server) PutAll(in map[string]string, out *Nothing) error {
	finished := make(chan struct{})
	s <- func(n *Node) {
		for key, value := range in {
			_, ok := n.data[key]
			if !ok {
				n.data[key] = value
			}
		}
		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (s Server) FindSuccessor(id *big.Int, reply *FindAnswer) error {
	finished := make(chan struct{})

	s <- func(n *Node) {

		if between(hashString(n.address), id, hashString(n.successor[0]), true) {
			*reply = FindAnswer{OK: true, Successor: n.successor[0]}
		} else {
			*reply = FindAnswer{OK: false, Successor: n.ClosePrecedingNode(id)}
		}

		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (s Server) Notify(might string, out *bool) error {
	finished := make(chan struct{})
	s <- func(n *Node) {

		if n.predecesor == "" || between(hashString(n.predecesor), hashString(might), hashString(n.address), false) {
			n.predecesor = might
			*out = true
		}

		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (s Server) LookUp(junk *Nothing, reply *[]string) error {
	finished := make(chan struct{})
	s <- func(n *Node) {

		var contain []string
		*reply = make([]string, len(n.successor)+1)
		contain = append(contain, n.predecesor)
		for i := range n.successor {
			contain = append(contain, n.successor[i])
		}
		copy(*reply, contain)
		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (s Server) Get(key string, reply *string) error {
	finished := make(chan struct{})
	s <- func(n *Node) {
		_, ok := n.data[key]
		if ok {
			*reply = n.data[key]
		}
		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (s Server) Delect(key string, reply *bool) error {
	finished := make(chan struct{})
	s <- func(n *Node) {
		_, ok := n.data[key]
		if ok {
			delete(n.data, key)
			*reply = true
		}
		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (s Server) Put(msg []string, reply *Nothing) error {
	finished := make(chan struct{})
	s <- func(n *Node) {
		n.data[msg[0]] = msg[1]
		// fmt.Printf("key : %d value : %d", hashString(msg[1]), hashString(msg[2]))
		finished <- struct{}{}
	}
	<-finished
	return nil
}

func server(address string, node *Node) {
	log.Printf("starting to listen %s\n", address)

	actor := startActor(node)
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

func startActor(node *Node) Server {
	ch := make(chan handler)
	state := node
	go func() {
		for f := range ch {
			f(state)
		}
	}()
	return ch
}

func call(address string, method string, request interface{}, reply interface{}) error {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Printf("rpc.DialHTTP: %v", err)
		return err
	}
	defer client.Close()

	if err = client.Call(method, request, reply); err != nil {
		log.Printf("client.Call: %v\n", err)
		return err
	}

	return nil
}
