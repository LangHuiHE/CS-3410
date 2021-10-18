package main

import (
	"fmt"
	"strconv"
)

// ACTION

func (state *State) CollectPrepare(voter int, proposal int, proposer *Proposer, reply int, time int) {
	if reply == 1 {
		proposer.preparevote[proposal].positive++
	} else {
		proposer.preparevote[proposal].negative++
	}
	proposer.preparevote[proposal].voted = append(proposer.preparevote[proposal].voted, voter)

	// check vote
	if proposer.preparevote[proposal].positive == state.majority {
		// start the accepte round
		var v, n int
		if state.nodes[proposer.number].na == 0 {
			v = proposer.number * 11111
			fmt.Printf("--> prepare round successful: %d proposing its own value %d\n", proposer.number, v)
		} else {
			v = state.nodes[proposer.number].va
			n = state.nodes[proposer.number].na
			fmt.Printf("--> prepare round successful: %d proposing discovered value %d sequence %d\n", proposer.number, v, n)
		}

		for i := range state.nodes {
			if i != 0 {
				key := Key{Type: MsgAcceptRequest, Time: time, Target: i}
				state.network[key] = "voter : " + strconv.Itoa(voter) + " n : " + strconv.Itoa(proposal) + " " + " v : " + strconv.Itoa(v) + " proposer: " + strconv.Itoa(proposer.number) + " \n"
			}
		}
		fmt.Printf("--> sent accept requests to all nodes from %d with value %d sequence %d\n", proposer.number, v, proposal)
	}

	if proposer.preparevote[proposal].negative == state.majority {
		// abandon and start a new prepare round
		fmt.Printf("--> accept round failed at %d, restarting\n", proposer.number)
		state.sendPrepare(proposer, time)
	}
}

func (state *State) CollectAccept(voter int, proposal int, proposer *Proposer, value int, reply int, time int) {
	if reply == 1 {
		proposer.acceptvote[proposal].positive++
	} else {
		proposer.acceptvote[proposal].negative++
	}
	proposer.acceptvote[proposal].voted = append(proposer.acceptvote[proposal].voted, voter)

	// check vote
	if proposer.acceptvote[proposal].positive == state.majority {
		fmt.Printf("--> accept round successful: %d detected consensus with value %d\n", proposer.number, value)
		// start the decide round
		for i := range state.nodes {
			if i > 0 {
				key := Key{Type: MsgDecideRequest, Time: time, Target: i}
				state.network[key] = "v: " + strconv.Itoa(value) + " \n"
			}
		}
		fmt.Printf("--> sent decide requests to all nodes from %d with value %d\n", proposer.number, value)
	}

	if proposer.acceptvote[proposal].negative == state.majority {
		// abandon and start a new prepare round
		fmt.Printf("--> accept round failed at %d, restarting\n", proposer.number)
		state.sendPrepare(proposer, time)
	}
}
