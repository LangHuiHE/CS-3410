package main

import (
	"bufio"
	"log"
	"os"
	"strings"
)

/*
node states:
proposer:

											-> (negative = majority) -> RESTART
	send prepare Request -> collect vote	->  (do nothing and keep waiting)
											-> (postive = majority) ->

1. Send Prepare Request / Waiting Prepare Response (Collecting Votes)
2. Recived Prepare Rqeuest / Reply Prepare Response
3.
4.

node actions:
1. Send out prepare requeset
2. Reply prepare request
3. Collecting votes
4.
*/

// message type
const (
	MsgPrepareRequest = iota
	MsgPrepareResponse
	MsgAcceptRequest
	MsgAcceptResponse
	MsgDecideRequest
)

type Key struct {
	Type   int
	Time   int
	Target int
}

type Node struct {
	np int // largest proposal seen in prepare
	na int // largest proposal seen in accept
	va int // value accepted for proposal Na
}

type Vote struct {
	positive int
	negative int
	voted    []int
}
type Proposer struct {
	restart     int
	number      int
	preparevote Vote
	acceptvote  Vote
}

// BIG state represente the whole simulator
type State struct {
	network  map[Key]string //record all the msg has sent to the network
	nodes    []Node
	proposer Proposer
	majority int
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	network := make(map[Key]string)
	state := State{network: network}
	for scanner.Scan() {
		line := scanner.Text()

		// trim comments
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}

		// ignore empty/comment-only lines
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}
		line += "\n"

		switch {
		case state.TryInitialize(line):

		case state.TrySendPrepare(line):

		case state.TryDeliverPrepareRequest(line):

		case state.TryDeliverPrepareResponse(line):

		case state.TryDeliverAcceptRequest(line):

		case state.TryDeliverAcceptResponse(line):

		case state.TryDeliverDecideRequeset(line):

		default:
			log.Fatalf("unknown line: %s\n", line)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("scanner failure: %v\n", err)
	}
}
