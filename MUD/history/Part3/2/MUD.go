package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"

	//"os"
	"net"
	"strings"
)

// Zone : struct
type Zone struct {
	ID    int
	Name  string
	Rooms []*Room
}

// Room struct
type Room struct {
	ID          int
	Zone        *Zone
	Name        string
	Description string
	Exits       [6]Exit
}

// Exit struct
type Exit struct {
	To          *Room
	Description string
}

// Direction : string to int
var Direction = map[string]int{
	"n": 0,
	"e": 1,
	"w": 2,
	"s": 3,
	"u": 4,
	"d": 5,
}

// ReDirection : numebr string
var ReDirection = map[int]string{
	0: "n",
	1: "e",
	2: "w",
	3: "s",
	4: "u",
	5: "d",
}

// Player struct
type Player struct {
	Name     string
	Outputs  chan OutputEvent
	Location *Room
	Login    bool
	Logout   bool
}

const listenAddress = ":3410"

// InputEvent strcut
type InputEvent struct {
	Player  *Player
	Command string
}

// OutputEvent struct
type OutputEvent struct {
	Mag string
}

// ALLROOM all the room
var ALLROOM = make(map[int]*Room)

// Part 1:
var allCommands = make(map[string]func(string, *Player))

var onlineplayers []*Player

// the MAKE bulid this type of data set in map and bulid the refrence

//func commandLoop(player Player) error {
//	initCommands()
//	cmdwhere("", player)
//	scanner := bufio.NewScanner(os.Stdin)
//	for scanner.Scan() {
//		line := scanner.Text()
//log.Printf("The line was: %s", line)

//		doCommand(line, player)
//	}
//	if err := scanner.Err(); err != nil {
//		return fmt.Errorf("in main command loop: %v", err)
//	}

//	return nil
//}

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

func initCommands() {
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

func doCommand(cmd string, player *Player) {
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

func cmdOnline(s string, player *Player) {
	player.Outputs <- OutputEvent{Mag: "Online players list:\n"}
	for index := range onlineplayers {
		player.Outputs <- OutputEvent{Mag: onlineplayers[index].Name + "\n"}
	}
	player.Outputs <- OutputEvent{Mag: "\n"}
}

func cmdShout(s string, player *Player) {
	words := strings.Fields(s)
	for index := range onlineplayers {
		if onlineplayers[index].Location.Zone.ID == player.Location.Zone.ID && onlineplayers[index].Name != player.Name {
			onlineplayers[index].Outputs <- OutputEvent{Mag: player.Name + " shout: " + strings.Join(words[1:], " ") + "\n"}
		}
	}
}

func cmdQuit(s string, player *Player) {
	player.Logout = true
}

func cmdMe(s string, player *Player) {
	player.Outputs <- OutputEvent{Mag: player.Name + "\n"}
}

func cmdLaugh(s string, player *Player) {
	var text string
	text = "HAHAHAHAAHA! You laught."
	player.Outputs <- OutputEvent{Mag: text}
}

func cmdTell(s string, player *Player) {
	words := strings.Fields(s)
	target := words[1]
	for index := range onlineplayers {
		if onlineplayers[index].Name == target {
			onlineplayers[index].Outputs <- OutputEvent{Mag: player.Name + " says: " + strings.Join(words[2:], " ") + "\n"}

		}
	}
}

func cmdSay(s string, player *Player) {
	words := strings.Fields(s)
	for index := range onlineplayers {
		if onlineplayers[index].Location.ID == player.Location.ID && onlineplayers[index].Name != player.Name {
			onlineplayers[index].Outputs <- OutputEvent{Mag: player.Name + " say: " + strings.Join(words[1:], " ") + "\n"}
		}
	}
}

func cmdLook(s string, player *Player) {

	var text string
	if player.checkexit(s) {
		lookNextLoca := player.Location.Exits[Direction[s]].Description
		text = lookNextLoca
		checkroom(player.Location.ID, player)
	} else {
		text = "There is nothing in that direcation.\n"
	}
	player.Outputs <- OutputEvent{Mag: text}
}

func cmdWhere(s string, player *Player) {
	var text string

	text = player.Location.Name + "\n\n"
	player.Outputs <- OutputEvent{Mag: text}

	text = player.Location.Description + "\n\n"
	player.Outputs <- OutputEvent{Mag: text}

	dirList := []string{}
	for d := 0; d < 6; d++ {
		if player.Location.Exits[d].To != nil {
			dirList = append(dirList, ReDirection[d])
		}
	}

	text = ("[ Exits: " + strings.Trim(fmt.Sprint(dirList), "[]") + " ]\n")
	player.Outputs <- OutputEvent{Mag: text}

	checkroom(player.Location.ID, player)

}

func cmdRecall(s string, player *Player) {
	text := "You take out a wishbone and pull apart. You see a light coming to you...\n"
	player.recall()
	player.Outputs <- OutputEvent{Mag: text}

	cmdWhere(s, player)
	enterroom(player.Location.ID, player)
}

func cmdNorth(s string, player *Player) {
	if player.checkexit("n") {
		oldLoca := player.Location.ID
		nextLoca := player.Location.Exits[Direction["n"]].To
		player.updatelocation(nextLoca)
		cmdWhere(s, player)
		leaveroom(oldLoca, player)
		enterroom(player.Location.ID, player)
	} else {
		text := "You can't go that direction!\n"
		player.Outputs <- OutputEvent{Mag: text}
	}
}

func cmdEast(s string, player *Player) {
	if player.checkexit("e") {
		oldLoca := player.Location.ID
		player.updatelocation(player.Location.Exits[Direction["e"]].To)
		cmdWhere(s, player)
		enterroom(player.Location.ID, player)
		leaveroom(oldLoca, player)
	} else {
		text := "You can't go that direction!\n"
		player.Outputs <- OutputEvent{Mag: text}
	}
}

func cmdWest(s string, player *Player) {
	if player.checkexit("w") {
		oldLoca := player.Location.ID
		player.updatelocation(player.Location.Exits[Direction["w"]].To)
		cmdWhere(s, player)
		enterroom(player.Location.ID, player)
		leaveroom(oldLoca, player)
	} else {
		text := "You can't go that direction!\n"
		player.Outputs <- OutputEvent{Mag: text}
	}
}

func cmdSouth(s string, player *Player) {
	if player.checkexit("s") {
		oldLoca := player.Location.ID
		player.updatelocation(player.Location.Exits[Direction["s"]].To)
		cmdWhere(s, player)
		enterroom(player.Location.ID, player)
		leaveroom(oldLoca, player)
	} else {
		text := "You can't go that direction!\n"
		player.Outputs <- OutputEvent{Mag: text}
	}
}

func cmdUp(s string, player *Player) {
	if player.checkexit("u") {
		oldLoca := player.Location.ID
		player.updatelocation(player.Location.Exits[Direction["u"]].To)
		cmdWhere(s, player)
		enterroom(player.Location.ID, player)
		leaveroom(oldLoca, player)
	} else {
		text := "You can't go that direction!\n"
		player.Outputs <- OutputEvent{Mag: text}
	}
}

func cmdDown(s string, player *Player) {
	if player.checkexit("d") {
		oldLoca := player.Location.ID
		player.updatelocation(player.Location.Exits[Direction["d"]].To)
		cmdWhere(s, player)
		enterroom(player.Location.ID, player)
		leaveroom(oldLoca, player)
	} else {
		text := "You can't go that direction!\n"
		player.Outputs <- OutputEvent{Mag: text}
	}
}

// Part 2:

func loadzone(transaction *sql.Tx) (map[int]*Zone, error) {

	var id int
	//var zone_id int
	var name string
	//var description string

	var allzone = make(map[int]*Zone)

	rows, err := transaction.Query(`SELECT id, name FROM zones ORDER BY id`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}

		newZone := Zone{id, name, nil}
		newPointer := &newZone
		allzone[id] = newPointer
	}

	return allzone, nil
}

func loadroom(Zone map[int]*Zone, transaction *sql.Tx) (map[int]*Room, error) {

	var roomID int
	var roomName string
	var roomDescription string
	var allroom = make(map[int]*Room)
	var zoneID int

	rows, err := transaction.Query(`SELECT id, zone_id, name, description FROM rooms`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&roomID, &zoneID, &roomName, &roomDescription); err != nil {
			return nil, err
		}

		zonePointer := Zone[zoneID]
		newRoom := Room{roomID, zonePointer, roomName, roomDescription, [6]Exit{}}
		allroom[roomID] = &newRoom
		Zone[zoneID].Rooms = []*Room{&newRoom}

	}
	return allroom, nil
}

func loadExit(Room map[int]*Room, transaction *sql.Tx) (map[int]*Room, error) {
	var fromRoom int
	var toRoom int
	var direction string
	var description string

	rows, err := transaction.Query(`SELECT * FROM exits`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&fromRoom, &toRoom, &direction, &description); err != nil {
			return nil, err
		}

		toPointer := Room[toRoom]
		newExit := Exit{toPointer, description}
		Room[fromRoom].Exits[Direction[direction]] = newExit
	}

	return Room, nil
}

func loadalldata(tx *sql.Tx) error {
	ALLzone, err := loadzone(tx)
	if err != nil {
		return err
	}

	ALLroom, err := loadroom(ALLzone, tx)

	ALLROOM, err = loadExit(ALLroom, tx)

	if err != nil {
		return err
	}

	//for key, value := range ALLROOM {
	//	fmt.Printf("ROOM: ", key, "pointer: ", value)
	//	fmt.Printf("\n")
	//}
	return nil
}

func opentransaction(database *sql.DB) error {
	tx, err := database.Begin()
	if err != nil {
		return fmt.Errorf("Starting Transaction: %v", err)
	}

	err = loadalldata(tx)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Loading world data: %v", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Transaction failed while loading world data : %v", err)
	}
	return nil
}

// Player methods:

func initplayer(name string, Outputs chan OutputEvent) *Player {
	player := Player{name, Outputs, ALLROOM[3001], true, false}
	return &player
}

func (player *Player) recall() {
	leaveroom(player.Location.ID, player)
	player.Location = ALLROOM[3001]
}

func (player *Player) updatelocation(newRoom *Room) {
	player.Location = newRoom
}

func (player *Player) checkexit(dir string) bool {
	numDir := Direction[dir]
	return player.Location.Exits[numDir].To != nil
}

// setup world
func setupWorld() error {
	// Part 2:
	// the path to the database--this could be an absolute path
	path := "world.db"
	options :=
		"?" + "_busy_timeout=10000" +
			"&" + "_foreign_keys=ON" +
			"&" + "_journal_mode=WAL" +
			"&" + "mode=rw" +
			"&" + "_synchronous=NORMAL"

	db, err := sql.Open("sqlite3", path+options)
	if err != nil {
		// handle the error here
		return err
	}

	defer db.Close()

	err = opentransaction(db)
	if err != nil {
		return err
	}
	return nil
}

func setup() error {
	initCommands()
	err := setupWorld()
	if err != nil {
		return err
	}
	return nil
}

func mainLoop(inputChannel chan InputEvent) {
	for event := range inputChannel {
		fmt.Printf("type: %s \n", event.Command)
		responseoutput(event)
	}
}

func responseoutput(event InputEvent) {
	// Part 1:
	doCommand(event.Command, event.Player)
	checkonline()
}

func checkonline() {
	for index := range onlineplayers {
		if onlineplayers[index].Login {
			for i := range onlineplayers {
				if i != index {
					onlineplayers[i].Outputs <- OutputEvent{Mag: "\n" + onlineplayers[index].Name + " is online now! \n"}
				}
			}
			onlineplayers[index].Login = false
		}

		if onlineplayers[index].Logout {
			for i := range onlineplayers {
				if i != index {
					onlineplayers[i].Outputs <- OutputEvent{Mag: "\n" + onlineplayers[index].Name + " just left the world!\n"}
				}
			}
			close(onlineplayers[index].Outputs)
			onlineplayers[index].Outputs = nil
			onlineplayers[index] = onlineplayers[len(onlineplayers)-1]
			onlineplayers[len(onlineplayers)-1] = nil
			onlineplayers = onlineplayers[:len(onlineplayers)-1]

		}
	}
}

func enterroom(room int, player *Player) {
	for index := range onlineplayers {
		if onlineplayers[index].Location.ID == room && onlineplayers[index].Name != player.Name {
			onlineplayers[index].Outputs <- OutputEvent{Mag: "\n" + player.Name + " enters the room!\n"}
		}
	}
}

func leaveroom(room int, player *Player) {
	for index := range onlineplayers {
		if onlineplayers[index].Location.ID == room && onlineplayers[index].Name != player.Name {
			onlineplayers[index].Outputs <- OutputEvent{Mag: "\n" + player.Name + " leaves the room!\n"}
		}
	}
}

func checkroom(room int, player *Player) {
	for index := range onlineplayers {
		if onlineplayers[index].Location.ID == room && onlineplayers[index].Name != player.Name {
			player.Outputs <- OutputEvent{Mag: "\n" + onlineplayers[index].Name + " is in the room!!\n"}
		}
	}
}
func listenForConnections(inputChannel chan InputEvent) {
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
	for scanner.Scan() {
		name := scanner.Text()
		player := initplayer(name, Outputs)
		onlineplayers = append(onlineplayers, player)

		//inputChannel <- InputEvent{Player: player, Command: "me"}
		inputChannel <- InputEvent{Player: player, Command: "where"}

		go sendoutdata(conn, player)

		//scanner = bufio.NewScanner(conn)
		for scanner.Scan() {
			line := scanner.Text()
			//fmt.Print(line)
			event := InputEvent{Player: player, Command: line}
			inputChannel <- event
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Connection error: %v", err)
		} else {
			log.Printf("Connection closed normally\n")
		}
	}
}

func sendoutdata(conn net.Conn, player *Player) {
	//n := 0
	for outevent := range player.Outputs {
		fmt.Fprint(conn, outevent.Mag)
		//fmt.Fprint(conn, n)
		//n++
	}

	if player.Logout {
		conn.Close()
	}
}

func main() {
	err := setup()
	if err != nil {
		log.Fatalf("Setting up: %v", err)
	}

	// connection part:
	inputChannel := make(chan InputEvent)
	go listenForConnections(inputChannel)
	mainLoop(inputChannel)
}
