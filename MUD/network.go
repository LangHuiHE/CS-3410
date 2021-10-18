package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"syscall"
)

const listenAddress = ":3410"

// MainLoop start to whole connection
func MainLoop(inputChannel chan InputEvent) {
	for event := range inputChannel {
		//fmt.Printf("type: %s \n", event.Command)
		DoCommand(event.Command, event.Player)
		checkonline()
	}
}

func checkonline() {
	for index := 0; index < len(Onlineplayers); index++ {
		if Onlineplayers[index].Login {
			for i := range Onlineplayers {
				if i != index {
					Onlineplayers[i].Outputs <- OutputEvent{Mag: "\n" + Onlineplayers[index].Name + " is online now! \n"}
				}
			}
			Onlineplayers[index].Login = false
		}

		if Onlineplayers[index].Logout {
			for i := range Onlineplayers {
				if i != index {
					Onlineplayers[i].Outputs <- OutputEvent{Mag: "\n" + Onlineplayers[index].Name + " just left the world!\n"}
				}
			}

			Onlineplayers[index].Outputs = nil
			log.Printf("The outgoing channel for the player: %s is nil \n", Onlineplayers[index].Name)

			Onlineplayers[index] = Onlineplayers[len(Onlineplayers)-1]
			Onlineplayers[len(Onlineplayers)-1] = nil
			Onlineplayers = Onlineplayers[:len(Onlineplayers)-1]
			log.Print("player removed\n")
		}
	}
}

// ListenForConnections go
func ListenForConnections(inputChannel chan InputEvent) {
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
		go handleConnection(conn, inputChannel)
	}
}

func handleConnection(conn net.Conn, inputChannel chan InputEvent) {

	fmt.Fprintf(conn, "What is your name?\n")
	Outputs := make(chan OutputEvent)
	scanner := bufio.NewScanner(conn)

	scanner.Scan()
	name := scanner.Text()
	p := Initplayer(name, Outputs)
	Onlineplayers = append(Onlineplayers, p)
	//inputChannel <- InputEvent{Player: player, Command: "me"}
	inputChannel <- InputEvent{Player: p, Command: "where"}
	go sendoutdata(conn, p)

	for {
		select {
		case <-Outputs:
			conn.Close()
			break
		default:
			err := connCheck(conn)
			if err != nil {
				p.Logout = true
				break
			}
			//scanner = bufio.NewScanner(conn)
			scanner.Scan()
			line := scanner.Text()
			//fmt.Print(line)
			event := InputEvent{Player: p, Command: line}
			inputChannel <- event
			if Outputs == nil {
				conn.Close()
				break
			}

			if err := scanner.Err(); err != nil {
				log.Printf("Connection error: %v", err)
			} else {
				// log.Printf("Connection closed normally\n")
			}
		}
	}
}

func connCheck(conn net.Conn) error {
	var sysErr error = nil
	rc, err := conn.(syscall.Conn).SyscallConn()
	if err != nil {
		return err
	}
	err = rc.Read(func(fd uintptr) bool {
		var buf []byte = []byte{0}
		n, _, err := syscall.Recvfrom(int(fd), buf, syscall.MSG_PEEK|syscall.MSG_DONTWAIT)
		switch {
		case n == 0 && err == nil:
			sysErr = io.EOF
		case err == syscall.EAGAIN || err == syscall.EWOULDBLOCK:
			sysErr = nil
		default:
			sysErr = err
		}
		return true
	})
	if err != nil {
		return err
	}

	return sysErr
}

func sendoutdata(conn net.Conn, p *Player) {
	if p.Outputs == nil {
		conn.Close()
	}
	for outevent := range p.Outputs {
		fmt.Fprint(conn, outevent.Mag)
	}
}
