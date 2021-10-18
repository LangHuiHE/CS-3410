package main

import (
	"fmt"
	"log"
	"strings"
)

// Part 1:
var allCommands = make(map[string]func(string, *Player))

func addCommand(cmd string, f func(string, *Player)) {
	for i := range cmd {
		if i == 0 {
			continue
		}
		prefix := cmd[:i]
		allCommands[prefix] = f
	}
	allCommands[cmd] = f
}

// InitCommands init
func InitCommands() {
	addCommand("online", cmdOnline)
	addCommand("laugh", cmdLaugh)
	addCommand("quit", cmdQuit)
	addCommand("say", cmdSay)
	addCommand("tell", cmdTell)
	addCommand("shout", cmdShout)

	addCommand("look", cmdLook)
	addCommand("where", cmdWhere)
	addCommand("me", cmdMe)

	addCommand("recall", cmdRecall)
	addCommand("north", cmdNorth)
	addCommand("east", cmdEast)
	addCommand("west", cmdWest)
	addCommand("south", cmdSouth)
	addCommand("up", cmdUp)
	addCommand("down", cmdDown)
}

// DoCommand do command
func DoCommand(cmd string, player *Player) {
	fmt.Printf("do command recives: %s \n", cmd)
	words := strings.Fields(cmd)
	if len(words) == 0 {
	} else {
		for i := 0; i < len(words); i++ {
			if f, exists := allCommands[strings.ToLower(words[i])]; exists {
				//fmt.Print(words)
				f(cmd, player)
			}
		}
	}
}

func cmdOnline(s string, p *Player) {
	p.Outputs <- OutputEvent{Mag: "Online players list:\n"}
	for index := range Onlineplayers {
		p.Outputs <- OutputEvent{Mag: Onlineplayers[index].Name + "\n"}
	}
	p.Outputs <- OutputEvent{Mag: "\n"}
}

func cmdShout(s string, p *Player) {
	words := strings.Fields(s)
	for index := range Onlineplayers {
		if Onlineplayers[index].Location.Zone.ID == p.Location.Zone.ID && Onlineplayers[index].Name != p.Name {
			Onlineplayers[index].Outputs <- OutputEvent{Mag: p.Name + " shout: " + strings.Join(words[1:], " ") + "\n"}
		}
	}
}

func cmdQuit(s string, player *Player) {
	close(player.Outputs)
	log.Printf("Closing the outgoing channel for the player: %s \n", player.Name)
	player.Logout = true
	checkonline()
}

func cmdMe(s string, player *Player) {
	player.Outputs <- OutputEvent{Mag: player.Name + "\n"}
}

func cmdLaugh(s string, player *Player) {
	var text string
	text = "HAHAHAHAAHA! You laught."
	player.Outputs <- OutputEvent{Mag: text}
}

func cmdTell(s string, p *Player) {
	words := strings.Fields(s)
	target := words[1]
	for index := range Onlineplayers {
		if Onlineplayers[index].Name == target {
			Onlineplayers[index].Outputs <- OutputEvent{Mag: p.Name + " says: " + strings.Join(words[2:], " ") + "\n"}

		}
	}
}

func cmdSay(s string, p *Player) {
	words := strings.Fields(s)
	for index := range Onlineplayers {
		if Onlineplayers[index].Location.ID == p.Location.ID && Onlineplayers[index].Name != p.Name {
			Onlineplayers[index].Outputs <- OutputEvent{Mag: p.Name + " say: " + strings.Join(words[1:], " ") + "\n"}
		}
	}
}

func cmdLook(s string, p *Player) {

	var text string
	if p.Checkexit(s) {
		lookNextLoca := p.Location.Exits[Direction[s]].Description
		text = lookNextLoca
		checkroom(p.Location.ID, p)
	} else {
		text = "There is nothing in that direcation.\n"
	}
	p.Outputs <- OutputEvent{Mag: text}
}

func cmdWhere(s string, p *Player) {
	var text string

	text = p.Location.Name + "\n\n"
	p.Outputs <- OutputEvent{Mag: text}

	text = p.Location.Description + "\n\n"
	p.Outputs <- OutputEvent{Mag: text}

	dirList := []string{}
	for d := 0; d < 6; d++ {
		if p.Location.Exits[d].To != nil {
			dirList = append(dirList, ReDirection[d])
		}
	}

	text = ("[ Exits: " + strings.Trim(fmt.Sprint(dirList), "[]") + " ]\n")
	p.Outputs <- OutputEvent{Mag: text}

	checkroom(p.Location.ID, p)

}

func cmdRecall(s string, p *Player) {
	text := "You take out a wishbone and pull apart. You see a light coming to you...\n"
	p.Recall()
	p.Outputs <- OutputEvent{Mag: text}

	cmdWhere(s, p)
	enterroom(p.Location.ID, p)
}

func cmdNorth(s string, p *Player) {
	if p.Checkexit("n") {
		oldLoca := p.Location.ID
		nextLoca := p.Location.Exits[Direction["n"]].To
		p.Updatelocation(nextLoca)
		cmdWhere(s, p)
		leaveroom(oldLoca, p)
		enterroom(p.Location.ID, p)
	} else {
		text := "You can't go that direction!\n"
		p.Outputs <- OutputEvent{Mag: text}
	}
}

func cmdEast(s string, p *Player) {
	if p.Checkexit("e") {
		oldLoca := p.Location.ID
		p.Updatelocation(p.Location.Exits[Direction["e"]].To)
		cmdWhere(s, p)
		enterroom(p.Location.ID, p)
		leaveroom(oldLoca, p)
	} else {
		text := "You can't go that direction!\n"
		p.Outputs <- OutputEvent{Mag: text}
	}
}

func cmdWest(s string, p *Player) {
	if p.Checkexit("w") {
		oldLoca := p.Location.ID
		p.Updatelocation(p.Location.Exits[Direction["w"]].To)
		cmdWhere(s, p)
		enterroom(p.Location.ID, p)
		leaveroom(oldLoca, p)
	} else {
		text := "You can't go that direction!\n"
		p.Outputs <- OutputEvent{Mag: text}
	}
}

func cmdSouth(s string, p *Player) {
	if p.Checkexit("s") {
		oldLoca := p.Location.ID
		p.Updatelocation(p.Location.Exits[Direction["s"]].To)
		cmdWhere(s, p)
		enterroom(p.Location.ID, p)
		leaveroom(oldLoca, p)
	} else {
		text := "You can't go that direction!\n"
		p.Outputs <- OutputEvent{Mag: text}
	}
}

func cmdUp(s string, p *Player) {
	if p.Checkexit("u") {
		oldLoca := p.Location.ID
		p.Updatelocation(p.Location.Exits[Direction["u"]].To)
		cmdWhere(s, p)
		enterroom(p.Location.ID, p)
		leaveroom(oldLoca, p)
	} else {
		text := "You can't go that direction!\n"
		p.Outputs <- OutputEvent{Mag: text}
	}
}

func cmdDown(s string, p *Player) {
	if p.Checkexit("d") {
		oldLoca := p.Location.ID
		p.Updatelocation(p.Location.Exits[Direction["d"]].To)
		cmdWhere(s, p)
		enterroom(p.Location.ID, p)
		leaveroom(oldLoca, p)
	} else {
		text := "You can't go that direction!\n"
		p.Outputs <- OutputEvent{Mag: text}
	}
}

func enterroom(room int, p *Player) {
	for index := range Onlineplayers {
		if Onlineplayers[index].Location.ID == room && Onlineplayers[index].Name != p.Name {
			Onlineplayers[index].Outputs <- OutputEvent{Mag: "\n" + p.Name + " enters the room!\n"}
		}
	}
}

func leaveroom(room int, p *Player) {
	for index := range Onlineplayers {
		if Onlineplayers[index].Location.ID == room && Onlineplayers[index].Name != p.Name {
			Onlineplayers[index].Outputs <- OutputEvent{Mag: "\n" + p.Name + " leaves the room!\n"}
		}
	}
}

func checkroom(room int, p *Player) {
	for index := range Onlineplayers {
		if Onlineplayers[index].Location.ID == room && Onlineplayers[index].Name != p.Name {
			p.Outputs <- OutputEvent{Mag: "\n" + Onlineplayers[index].Name + " is in the room!!\n"}
		}
	}
}
