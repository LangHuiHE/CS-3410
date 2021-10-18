package main

import "fmt"

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

	fmt.Printf("Done Initialize %d Nodes(node 0 is not used)\n", len(state.nodes))
	fmt.Printf("Majority is %d\n", state.majority)
	return true
}

func (state *State) TrySendPrepare(line string) bool {
	var time, from int
	n, err := fmt.Sscanf(line, "at %d send prepare request from %d\n", &time, &from)
	if err != nil || n != 2 {
		return false
	}

	state.proposer.number = from

	state.sendPrepare(time)
	fmt.Print("Done Send Prepare\n")
	return true
}

func (state *State) TryDeliverPrepareRequest(line string) bool {
	var time, fromTime, target int
	n, err := fmt.Sscanf(line, "at %d deliver prepare request message to %d from time %d\n", &time, &target, &fromTime)
	if err != nil || n != 3 {
		//log.Fatal("TryDeliverPrepareRequest read line\n")
		return false
	}

	lookUp := Key{Type: 0, Time: fromTime, Target: target}

	var proposal, reply int
	count, err := fmt.Sscanf(state.network[lookUp], "n: %d\n", &proposal)
	if err != nil || count != 1 {
		//log.Fatal("TryDeliverPrepareRequest read context\n")
		return false
	}
	if proposal > state.nodes[target].np {
		state.nodes[target].np = proposal
		reply = 1
	} else {
		reply = -1
	}

	state.PrepareResponse(target, time, proposal, state.nodes[target].na, state.nodes[target].va, reply)

	fmt.Print("Done Prepare Request\n")
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
	var voter, proposal, na, va, reply int

	count, err := fmt.Sscanf(state.network[key], "voter: %d n: %d na: %d va: %d reply: %d\n", &voter, &proposal, &na, &va, &reply)
	if err != nil || count != 5 {
		//log.Fatal("TryDeliverPrepareResponse read context\n")
		return false
	}

	// check dupulicate massage
	for _, b := range state.proposer.preparevote.voted {
		if b == voter {
			fmt.Printf("%s is dupilicated\n", line)
			return true
		}
	}

	// proposer collect vote
	state.CollectPrepare(voter, proposal, reply, time)
	fmt.Print("Done Prepare Response\n")
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
	var proposal, v, reply int
	count, err := fmt.Sscanf(state.network[key], "n: %d v: %d\n", &proposal, &v)
	if err != nil || count != 2 {
		//log.Fatal("TryDeliverAcceptRequest read context\n")
		return false
	}

	if proposal >= state.nodes[target].np {
		state.nodes[target].na = proposal
		state.nodes[target].va = v
		reply = 1
	} else {
		reply = -1
	}
	state.AcceptResponse(time, proposal, reply, target)
	fmt.Print("Done Accept Request\n")
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

	var acceptor, reply, proposal int
	count, err := fmt.Sscanf(state.network[key], "acceptor: %d n: %d reply: %d\n", &acceptor, &proposal, &reply)
	if err != nil || count != 3 {
		//log.Fatal("TryDeliverAcceptResponse read context\n")
		return false
	}

	// check dupulicate massage
	for _, b := range state.proposer.acceptvote.voted {
		if b == acceptor {
			fmt.Printf("%s is duplicated\n", line)
			return true
		}
	}
	state.CollectAccept(acceptor, proposal, reply, time)
	fmt.Print("Done Accept Response\n")
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

	var v int
	count, err := fmt.Sscanf(state.network[key], "v: %d\n", &v)
	if err != nil || count != 1 {
		//log.Fatal("TryDeliverDecideRequeset read context\n")
		return false
	}
	state.nodes[target].va = v
	fmt.Printf("node: %d {%+v} \n", target, state.nodes[target])
	fmt.Print("Done Decide Requeset\n")
	return true
}
