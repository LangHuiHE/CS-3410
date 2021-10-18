package main

// InputEvent strcut
type InputEvent struct {
	Player  *Player
	Command string
}

// Player struct
type Player struct {
	Name     string
	Outputs  chan OutputEvent
	Location *Room
	Login    bool
	Logout   bool
}

// Onlineplayers : contains all the player online
var Onlineplayers []*Player

// Player methods:

// Initplayer : return a new player
func Initplayer(name string, Outputs chan OutputEvent) *Player {
	player := Player{name, Outputs, ALLROOM[3001], true, false}
	return &player
}

// Recall send player back to room 3001
func (player *Player) Recall() {
	player.Location = ALLROOM[3001]
}

// Updatelocation update player location
func (player *Player) Updatelocation(newRoom *Room) {
	player.Location = newRoom
}

// Checkexit check the string is in the option
func (player *Player) Checkexit(dir string) bool {
	numDir := Direction[dir]
	return player.Location.Exits[numDir].To != nil
}
