package main

import (
	"fmt"
	"strconv"
)

func (state *State) TryInitialize(line string) bool {
	var size int
	n, err := fmt.Sscanf(line, "initialize %d nodes\n", &size)
	if err != nil || n != 1 || size < 3 || size > 9 {
		return false
	}

	state.nodes = make([]Node, size+1)

	if size%2 != 0 {
		state.majority = (size + 1) / 2
	} else {
		state.majority = size/2 + 1
	}

	fmt.Printf("--> initialized %d nodes\n", len(state.nodes)-1)

	return true
}

func (state *State) TrySendPrepare(line string) bool {
	var time, from int
	n, err := fmt.Sscanf(line, "at %d send prepare request from %d\n", &time, &from)
	if err != nil || n != 2 {
		return false
	}

	if val, ok := state.proposer[from]; ok {
		state.sendPrepare(val, time)
	} else {
		newProposer := Proposer{restart: -1, number: from, preparevote: make(map[int]*Vote), acceptvote: make(map[int]*Vote), acceptrequest: make(map[int]*Vote)}
		state.proposer[from] = &newProposer
		state.sendPrepare(&newProposer, time)
	}
	return true
}

func (state *State) sendPrepare(proposer *Proposer, time int) {
	proposer.restart++
	proposal := 5000 + proposer.number + proposer.restart*10
	var pv = Vote{positive: 0, negative: 0, voted: []int{}}
	var av = Vote{positive: 0, negative: 0, voted: []int{}}

	proposer.preparevote[proposal] = &pv
	proposer.acceptvote[proposal] = &av

	for i := range state.nodes {
		if i > 0 {
			key := Key{Type: MsgPrepareRequest, Time: time, Target: i}
			state.network[key] = "n: " + strconv.Itoa(proposal) + " proposer: " + strconv.Itoa(proposer.number) + "\n"
		}
	}
	fmt.Printf("--> sent prepare requests to all nodes from %d with sequence %d\n", proposer.number, proposal)
}

func (state *State) TryDeliverPrepareRequest(line string) bool {
	var time, fromTime, target int
	n, err := fmt.Sscanf(line, "at %d deliver prepare request message to %d from time %d\n", &time, &target, &fromTime)
	if err != nil || n != 3 {
		//log.Fatal("TryDeliverPrepareRequest read line\n")
		return false
	}

	lookUp := Key{Type: 0, Time: fromTime, Target: target}

	var proposal, reply, proposer int
	count, err := fmt.Sscanf(state.network[lookUp], "n: %d proposer: %d\n", &proposal, &proposer)
	if err != nil || count != 2 {
		//log.Fatal("TryDeliverPrepareRequest read context\n")
		return false
	}

	if proposal > state.nodes[target].np {
		state.nodes[target].np = proposal
		reply = 1
	} else {
		reply = -1
	}

	if state.nodes[target].va == 0 {
		fmt.Printf("--> prepare request from %d sequence %d accepted by %d with no value\n", proposer, proposal, target)
	} else {
		fmt.Printf("--> prepare request from %d sequence %d accepted by %d with value %d sequence %d\n", proposer, proposal, target, state.nodes[target].va, state.nodes[target].na)
	}

	key := Key{Type: MsgPrepareResponse, Time: time, Target: proposer}
	state.network[key] = "voter: " + strconv.Itoa(target) + " " + "n: " + strconv.Itoa(proposal) + " " + "na: " + strconv.Itoa(state.nodes[target].na) + " " + "va: " + strconv.Itoa(state.nodes[target].va) + " " + "reply: " + strconv.Itoa(reply) + " proposer: " + strconv.Itoa(proposer) + "\n"
	return true
}

func (state *State) TryDeliverPrepareResponse(line string) bool {
	var time, fromTime, target int
	n, err := fmt.Sscanf(line, "at %d deliver prepare response message to %d from time %d\n", &time, &target, &fromTime)
	if err != nil || n != 3 {
		//log.Fatal("TryDeliverPrepareResponse read line \n")
		return false
	}

	key := Key{Type: 1, Time: fromTime, Target: target}

	var voter, proposal, na, va, reply, proposerNum int

	count, err := fmt.Sscanf(state.network[key], "voter: %d n: %d na: %d va: %d reply: %d proposer: %d\n", &voter, &proposal, &na, &va, &reply, &proposerNum)
	if err != nil || count != 6 {
		//log.Fatal("TryDeliverPrepareResponse read context\n")
		return false
	}

	proposer := state.proposer[proposerNum]

	// check dupulicate massage
	for _, b := range proposer.preparevote[proposal].voted {
		if b == voter {
			fmt.Printf("--> prepare response from %d sequence %d ignored as a duplicate by %d\n", voter, proposal, proposer.number)
			return true
		}
	}

	if reply == 1 {
		if va == 0 {
			fmt.Printf("--> positive prepare response from %d sequence %d recorded by %d with no value\n", voter, proposal, proposer.number)
		} else {
			if state.nodes[voter].va > state.nodes[proposer.number].va {
				state.nodes[proposer.number].va = state.nodes[voter].va
			}
			if state.nodes[voter].na > state.nodes[proposer.number].na {
				state.nodes[proposer.number].na = state.nodes[voter].na
			}
			fmt.Printf("--> positive prepare response from %d sequence %d recorded by %d with value %d sequence %d\n", voter, proposal, proposer.number, va, na)
		}
	}
	if reply == -1 {
		if va == 0 {
			fmt.Printf("--> negative prepare response from %d sequence %d recorded by %d with no value\n", voter, proposal, proposer.number)
		} else {
			if state.nodes[voter].va > state.nodes[proposer.number].va {
				state.nodes[proposer.number].va = state.nodes[voter].va
			}
			if state.nodes[voter].na > state.nodes[proposer.number].na {
				state.nodes[proposer.number].na = state.nodes[voter].na
			}
			fmt.Printf("--> negative prepare response from %d sequence %d recorded by %d with value %d sequence %d\n", voter, proposal, proposer.number, va, na)
		}
	}

	// proposer collect vote
	if len(proposer.preparevote[proposal].voted) == state.majority {
		fmt.Printf("--> valid prepare vote ignored by %d because round is already resolved\n", proposer.number)
		return true
	}

	state.CollectPrepare(voter, proposal, proposer, reply, time)

	return true
}

func (state *State) TryDeliverAcceptRequest(line string) bool {
	var time, fromTime, target int
	n, err := fmt.Sscanf(line, "at %d deliver accept request message to %d from time %d\n", &time, &target, &fromTime)
	if err != nil || n != 3 {
		//log.Fatal("TryDeliverAcceptRequest read line\n")
		return false
	}

	key := Key{Type: MsgAcceptRequest, Time: fromTime, Target: target}

	var proposal, v, reply, voter, proposerNum int
	count, err := fmt.Sscanf(state.network[key], "voter : %d n : %d v : %d proposer: %d\n", &voter, &proposal, &v, &proposerNum)
	if err != nil || count != 4 {
		//log.Fatal("TryDeliverAcceptRequest read context\n")
		return false
	}

	proposer := state.proposer[proposerNum]

	new := Vote{positive: 0, negative: 0, voted: []int{}}
	x := &new
	if val, ok := proposer.acceptrequest[proposal]; ok {
		x = val
	} else {
		proposer.acceptrequest[proposal] = x
	}

	if proposal >= state.nodes[target].np {
		state.nodes[target].na = proposal
		state.nodes[target].va = v
		reply = 1
		x.positive++
		fmt.Printf("--> accept request from %d with value %d sequence %d accepted by %d\n", proposer.number, v, proposal, target)
	} else {
		reply = -1
		x.negative++
		fmt.Printf("--> accept request from %d with value %d sequence %d rejected by %d\n", proposer.number, v, proposal, target)
	}

	responceKey := Key{Type: MsgAcceptResponse, Time: time, Target: proposer.number}
	state.network[responceKey] = "acceptor: " + strconv.Itoa(target) + " n: " + strconv.Itoa(proposal) + " " + " v: " + strconv.Itoa(v) + " " + "reply: " + strconv.Itoa(reply) + " proposer: " + strconv.Itoa(proposer.number) + " \n"

	if x.positive == state.majority {
		fmt.Print("--> note: consensus has been achieved\n")
	}
	return true
}

func (state *State) TryDeliverAcceptResponse(line string) bool {
	var time, fromTime, target int
	n, err := fmt.Sscanf(line, "at %d deliver accept response message to %d from time %d\n", &time, &target, &fromTime)
	if err != nil || n != 3 {
		//log.Fatal("TryDeliverAcceptResponse read line\n")
		return false
	}
	key := Key{Type: MsgAcceptResponse, Time: fromTime, Target: target}

	var acceptor, reply, proposal, value, proposerNum int
	count, err := fmt.Sscanf(state.network[key], "acceptor: %d n: %d v: %d reply: %d proposer: %d\n", &acceptor, &proposal, &value, &reply, &proposerNum)
	if err != nil || count != 5 {
		//log.Fatal("TryDeliverAcceptResponse read context\n")
		return false
	}

	proposer := state.proposer[proposerNum]

	// check dupulicate massage
	for _, b := range proposer.acceptvote[proposal].voted {
		if b == acceptor {
			fmt.Printf("--> accept response from %d sequence %d ignored as duplicate by %d", acceptor, proposal, proposer.number)
			return true
		}
	}
	if proposer.acceptvote[proposal].negative == state.majority {
		fmt.Printf("--> accept response from %d sequence %d from the past ignored by %d\n", acceptor, proposal, proposer.number)
	} else {
		if reply == 1 {
			fmt.Printf("--> positive accept response from %d sequence %d recorded by %d\n", acceptor, proposal, proposer.number)
		} else {
			fmt.Printf("--> negative accept response from %d sequence %d recorded by %d\n", acceptor, proposal, proposer.number)
		}

		if len(proposer.acceptvote[proposal].voted) == state.majority {
			fmt.Printf("--> valid accept vote ignored by %d because round is already resolved\n", proposer.number)
			return true
		}

		state.CollectAccept(acceptor, proposal, proposer, value, reply, time)
	}
	return true
}

func (state *State) TryDeliverDecideRequeset(line string) bool {
	var time, fromTime, target int
	n, err := fmt.Sscanf(line, "at %d deliver decide request message to %d from time %d\n", &time, &target, &fromTime)
	if err != nil || n != 3 {
		//log.Fatal("TryDeliverDecideRequeset read line\n")
		return false
	}
	key := Key{Type: MsgDecideRequest, Time: fromTime, Target: target}

	var value int
	count, err := fmt.Sscanf(state.network[key], "v: %d\n", &value)
	if err != nil || count != 1 {
		//log.Fatal("TryDeliverDecideRequeset read context\n")
		return false
	}
	state.nodes[target].va = value
	fmt.Printf("--> recording consensus value %d at %d\n", value, target)
	return true
}
