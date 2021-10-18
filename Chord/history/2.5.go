package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
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
	finger      [161]string
	successor   [lenSuccessor]string
	predecessor string
}

func main() {
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
