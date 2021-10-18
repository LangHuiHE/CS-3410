package main

import (
	"fmt"
	"strconv"
)

// ACTION

func (state *State) CollectPrepare(voter int, proposal int, reply int, time int) {
	if reply == 1 {
		state.proposer.preparevote.positive++
	} else {
		state.proposer.preparevote.negative++
	}
	state.proposer.preparevote.voted = append(state.proposer.preparevote.voted, voter)
	fmt.Printf("current Prepare vote count postive: %d negative: %d\n", state.proposer.preparevote.positive, state.proposer.preparevote.negative)
	fmt.Printf("Prepare voted node: %+q\n", state.proposer.preparevote.voted)

	// check vote
	if state.proposer.preparevote.positive == state.majority {
		// start the accepte round
		for i := range state.nodes {
			if i != 0 {
				key := Key{Type: MsgAcceptRequest, Time: time, Target: i}
				n := proposal
				v := state.proposer.number * 11111
				state.nodes[state.proposer.number].va = v
				state.network[key] = "n: " + strconv.Itoa(n) + " " + "v: " + strconv.Itoa(v) + " \n"
				fmt.Printf("Finished voting Send message to %d on network : %s\n", i, state.network[key])
			}
		}
		state.proposer.preparevote.positive = 0
		state.proposer.preparevote.negative = 0
		state.proposer.preparevote.voted = nil
		fmt.Printf("Finshed Prepare voting Reset vote count; postive: %d negative: %d voted: %+q\n", state.proposer.preparevote.positive, state.proposer.preparevote.negative, state.proposer.preparevote.voted)
	}

	if state.proposer.preparevote.negative == state.majority || state.proposer.preparevote.positive+state.proposer.preparevote.negative == len(state.nodes) {
		// abandon and start a new prepare round
		state.sendPrepare(time)
	}
}

func (state *State) sendPrepare(time int) {
	state.proposer.preparevote.positive = 0
	state.proposer.preparevote.negative = 0
	state.proposer.preparevote.voted = nil
	state.proposer.acceptvote.positive = 0
	state.proposer.acceptvote.negative = 0
	state.proposer.acceptvote.voted = nil
	fmt.Printf("Prepare voting Reset vote count; postive: %d negative: %d voted: %+q\n", state.proposer.preparevote.positive, state.proposer.preparevote.negative, state.proposer.preparevote.voted)

	for i := range state.nodes {
		if i > 0 {
			key := Key{Type: MsgPrepareRequest, Time: time, Target: i}
			n := 5000 + state.proposer.number + state.proposer.restart*10
			state.network[key] = "n: " + strconv.Itoa(n) + "\n"
			fmt.Printf("Send Prepare to %d on network : %s\n", i, state.network[key])
		}
	}
}

func (state *State) PrepareResponse(voter int, time int, n int, na int, va int, reply int) {
	key := Key{Type: MsgPrepareResponse, Time: time, Target: state.proposer.number}
	state.network[key] = "voter: " + strconv.Itoa(voter) + " " + "n: " + strconv.Itoa(n) + " " + "na: " + strconv.Itoa(na) + " " + "va: " + strconv.Itoa(va) + " " + "reply: " + strconv.Itoa(reply) + "\n"
	fmt.Printf("Send PreparePesponse to %d on network : %s\n", state.proposer.number, state.network[key])
}

func (state *State) AcceptResponse(time int, proposal int, reply int, acceptor int) {
	key := Key{Type: MsgAcceptResponse, Time: time, Target: state.proposer.number}
	state.network[key] = "acceptor: " + strconv.Itoa(acceptor) + " n: " + strconv.Itoa(proposal) + " " + "reply: " + strconv.Itoa(reply) + " \n"
	fmt.Printf("Send AcceptResponse to %d on network : %s\n", state.proposer.number, state.network[key])
}

func (state *State) CollectAccept(voter int, proposal int, reply int, time int) {
	if reply == 1 {
		state.proposer.acceptvote.positive++
	} else {
		state.proposer.acceptvote.negative++
	}
	state.proposer.acceptvote.voted = append(state.proposer.acceptvote.voted, voter)
	fmt.Printf("current Accept vote count postive: %d negative: %d\n", state.proposer.acceptvote.positive, state.proposer.acceptvote.negative)
	fmt.Printf("Accept voted node: %+q\n", state.proposer.acceptvote.voted)

	// check vote
	if state.proposer.acceptvote.positive == state.majority {
		// start the decide round
		for i := range state.nodes {
			if i > 0 {
				key := Key{Type: MsgDecideRequest, Time: time, Target: i}
				state.network[key] = "v: " + strconv.Itoa(state.nodes[state.proposer.number].va) + " \n"
				fmt.Printf("Finished Accept voting Send message to %d on network : %s\n", i, state.network[key])
			}
		}
		state.proposer.acceptvote.positive = 0
		state.proposer.acceptvote.negative = 0
		state.proposer.acceptvote.voted = nil
		fmt.Printf("Finished Accept vote count Reset vote count; postive: %d negative: %d voted node: %+q\n", state.proposer.acceptvote.positive, state.proposer.acceptvote.negative, state.proposer.acceptvote.voted)
	}

	if state.proposer.acceptvote.negative == state.majority || state.proposer.acceptvote.positive+state.proposer.acceptvote.negative == len(state.nodes) {
		// abandon and start a new prepare round
		state.sendPrepare(time)
	}
}
