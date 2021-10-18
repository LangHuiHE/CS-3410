package main

import (
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net"
	"reflect"
	"time"
)

var defaultHost = getLocalAddress()
var stop = make(chan struct{})

const (
	defaultPort     = "3410"
	sizeOfsuccessor = 3
	sizeOffinger    = 161
	maxSteps        = 3
)

type Node struct {
	id         *big.Int
	address    string
	successor  []string
	predecesor string
	finger     []string
	data       map[string]string
	lock       bool
}

type handler func(*Node)
type Server chan<- handler

type Nothing struct{}

func getLocalAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

// stabilize function
func (node *Node) stabilize() {
	//log.Print("run stabilize\n")
	var junk Nothing
	var reply []string

	checksum := node.successor

	for {
		if err := call(node.successor[0], "Server.LookUp", &junk, &reply); err != nil {
			if len(node.successor) > 1 {
				node.successor = node.successor[1:]
			} else {
				node.successor[0] = node.address
			}
		} else {
			break
		}
	}

	x := reply[0]
	if x != "" && between(hashString(node.address), hashString(x), hashString(node.successor[0]), false) {
		node.successor = append([]string{x}, node.successor...)
	}

	var newList []string
	newList = append(newList, node.successor[0])
	i := 1
	if reply[1] == newList[0] {
		i = 2
	}
	for i < len(reply) && len(newList) < sizeOfsuccessor {
		newList = append(newList, reply[i])
		i++
	}
	if !reflect.DeepEqual(node.successor, newList) && newList != nil {
		node.successor = newList
	}

	// notify()
	var updated bool
	if err := call(node.successor[0], "Server.Notify", node.address, &updated); err != nil {
		log.Printf("calling Server.Notify: %v\n", err)
	}

	if !reflect.DeepEqual(checksum, node.successor) || updated {
		log.Printf("\nnode has been updated\n----------\nNode Info:\nid: %s\naddress: %s\nsuccessor: %s\npredecessor: %s\n", node.id.String()[:4], node.address, node.successor, node.predecesor)
		fmt.Print("----------\n")
	}
}

// Random generate string for input
const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

func main() {
loop:
	for {
		select {
		case <-stop:
			break loop
		default:
			fmt.Printf("Local address: %s\n", getLocalAddress())
			fmt.Print("progrem is runing.\n")
			shell()
		}
	}
	fmt.Print("program is closing\n")
}
