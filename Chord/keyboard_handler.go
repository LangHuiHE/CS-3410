package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func shell() {
	log.Printf("Starting interactive shell")
	log.Printf("Commands are: \n\thelp\n\tquit\n\tport <n>\n\tcreate\n\tjoin <address>\n\tput <key> <value>\n\tputrandom <n>\n\tget <key>\n\tdelete <key>\n\tdump\n\tdumpkey <key>\n\tdumpaddr <address>\n\tdumpall\n\n")

	node := Node{id: hashString(net.JoinHostPort(defaultHost, defaultPort)), address: net.JoinHostPort(defaultHost, defaultPort), successor: []string{}, finger: make([]string, sizeOffinger), data: make(map[string]string), lock: false}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		parts := strings.Fields(line)

		if len(parts) == 0 {
			continue
		}
		switch parts[0] {

		case "dumpall":

		case "dumpaddr":

		case "dumpkey":

		case "dump":
			log.Printf("\n----------\nNode Info:\nid: %s\naddress: %s\nsuccessor: %s\npredecessor: %s\n", node.id.String()[:4], node.address, node.successor, node.predecesor)
			fmt.Print("node data:\n")
			fmt.Print("----------\n")
			if len(node.data) == 0 {
				fmt.Print("<-empty->\n")
			} else {

				displayCount := 0
				if len(parts) > 1 {
					i, err := strconv.Atoi(parts[1])
					if err == nil {
						displayCount = i
					}
				}
				var highest, lowest, count int

				for key, value := range node.data {
					hash := hashString(key)
					s := hash.String()
					seven_s := s[len(s)-7:]
					num, err := strconv.Atoi(seven_s)
					if err != nil {
						log.Fatalf("Converting key to hash number: %v", err)
					}
					if lowest == 0 {
						lowest = num
						highest = num
					}
					if num < lowest {
						lowest = num
					}
					if num > highest {
						highest = num
					}
					if displayCount == 0 {
						fmt.Printf("%s\tkey: %-30s\tvalue: %-30s\n", seven_s, key, value)
					}
					if count < displayCount {
						fmt.Printf("%s\tkey: %-30s\tvalue: %-30s\n", seven_s, key, value)
					}
					count++
				}
				if displayCount == 0 {
					fmt.Print("display all records\n")

				} else {
					fmt.Printf("only display first %d records\n", displayCount)
				}
				fmt.Printf("key range: %d ~ %d\ntotal key count: %d\n", lowest, highest, count)
			}
			fmt.Print("----------\n")

		case "delete":
			if len(parts) < 2 {
				log.Printf("you must specify the key for delection\n")
				continue
			}

			targetAddress := node.address
			key := parts[1]
			// find
			find := node.Find(hashString(key), node.address)
			if find == "ERROR" {
				//log.Printf("Key: %s Value: %s can't be stored plase try again\n", parts[1], parts[2])

			} else {
				targetAddress = find
			}

			var finished bool

			if err := call(targetAddress, "Server.Delect", key, &finished); err != nil {
				log.Fatalf("calling Server.Delect: %v", err)
			}
			if finished {
				log.Printf("\nDelete\n--------\naddress: %s %s \n--------\n", targetAddress, key)
			} else {
				fmt.Printf("key : %s does not exsist\n", key)
			}

		case "get":
			if len(parts) < 2 {
				log.Printf("you must specify the key for get\n")
				continue
			}

			var reply *string
			key := parts[1]
			targetAddress := node.address

			// find
			find := node.Find(hashString(key), node.address)
			if find == "ERROR" {
				//log.Printf("Key: %s Value: %s can't be stored plase try again\n", parts[1], parts[2])

			} else {
				targetAddress = find
			}

			if err := call(targetAddress, "Server.Get", key, &reply); err != nil {
				log.Fatalf("calling Server.Get: %v", err)
			}
			word := &reply
			if **word == "" {
				fmt.Printf("key : %s does not exsist\n", key)
			} else {
				log.Printf("\nGet\n--------\naddress: %s %s - %s\n--------\n", targetAddress, key, **word)
			}

		case "putrandom":
			if len(parts) != 2 {
				log.Printf("you must specify how many random text to put in\n")
				continue
			}
			number, err := strconv.Atoi(parts[1])
			if err != nil {
				continue
			}

			for i := 0; i < number; {
				var junk Nothing
				var randomtext []string
				randomkey := String(rand.Intn(26) + 1)

				randomtext = append(randomtext, randomkey)
				randomtext = append(randomtext, "junk")
				_, ok := node.data[randomkey]
				if !ok {
					i++
				}
				if err := call(node.address, "Server.Put", randomtext, &junk); err != nil {
					log.Fatalf("calling Server.Put: %v\n", err)
				}

			}

		case "put":
			if len(parts) < 2 {
				fmt.Print("you must specify a message to put in\n")
				continue
			}
			if len(parts) < 3 {
				fmt.Print("you should specify the key-value pair\n")
				continue
			}

			putAddress := node.address

			if len(parts) == 4 {
				putAddress = parts[3]
			}

			// find
			reply := node.Find(hashString(parts[1]), node.address)
			if reply == "ERROR" {
				//log.Printf("Key: %s Value: %s can't be stored plase try again\n", parts[1], parts[2])

			} else {
				putAddress = reply
			}

			var junk Nothing
			//log.Printf("Put to : %s\n", putAddress)
			if err := call(putAddress, "Server.Put", []string{parts[1], parts[2]}, &junk); err != nil {
				log.Fatalf("calling Server.Put: %v\n", err)
			} else {
				log.Printf("Put Success Key: %s Value: %s address: %s\n", parts[1], parts[2], putAddress)
			}

		case "quit":
			var junk Nothing
			if err := call(node.successor[0], "Server.PutAll", node.data, &junk); err != nil {
				log.Fatalf("calling Server.PutAll: %v\n", err)
			}
			close(stop)
			return

		case "join":
			if !node.lock {
				if len(parts) == 2 {
					var reply FindAnswer
					if err := call(parts[1], "Server.FindSuccessor", node.id, &reply); err != nil {
						log.Fatalf("calling Server.FindSuccessor: %v\n", err)
					}
					if reply.OK {
						node.successor = append(node.successor, reply.Successor)
					} else {
						node.successor = append(node.successor, parts[1])
					}
					var newMap map[string]string
					if err := call(node.successor[0], "Server.GetAll", node.address, &newMap); err != nil {
						log.Fatalf("calling Server.GetAll: %v\n", err)
					}

					node.data = newMap

					node.lock = true
					go server(node.address, &node)

					go func() {
						for {
							time.Sleep(1 * time.Second)
							node.stabilize()
							node.CheckPredecessor()
							node.FixFingers()
						}
					}()

				} else {
					fmt.Print("Plase entry the address for listening\n")
					continue
				}
			} else {
				fmt.Print("ERROR: Can't join a ring\n")
				continue
			}

		case "create":
			if !node.lock {
				node.successor = append(node.successor, node.address)
				node.lock = true

				go server(node.address, &node)

				go func() {
					for {
						time.Sleep(1 * time.Second)
						node.stabilize()
						node.CheckPredecessor()
						node.FixFingers()
					}
				}()

			} else {
				fmt.Print("ERROR: Can't create a ring\n")
				continue
			}

		case "port":
			var port string
			if len(parts) == 2 {
				if !node.lock {
					port = parts[1]
					node.address = net.JoinHostPort(defaultHost, port)
					node.id = hashString(node.address)
				} else {
					fmt.Print("ERROR: Can't no change the port is listening\n")
					continue
				}
			}
			fmt.Printf("The address for listening is : %s\n", node.address)
			continue
		case "help":
			fallthrough

		default:
			log.Printf("Only those commands are supported: \n\thelp\n\tquit\n\tport <n>\n\tcreate\n\tjoin <address>\n\tput <key> <value>\n\tputrandom <n>\n\tget <key>\n\tdelete <key>\n\tdump\n\tdumpkey <key>\n\tdumpaddr <address>\n\tdumpall\n\n")
		}
	}
}
