package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"fmt"
	"log"
	"bufio"
	"os"
	"strings"
	"net"
)

// Struct :

type Zone struct {
    ID    int 
    Name  string
    Rooms []*Room
}


type Room struct {
    ID          int 
    Zone        *Zone
    Name        string
    Description string
    Exits       [6]Exit
}

type Exit struct {
    To          *Room
    Description string
}

var Direction = map[string]int {
	"n" : 0,
	"e" : 1,
	"w" : 2,
	"s" : 3,
	"u" : 4,
	"d" : 5,
}
var ReDirection = map[int]string {
	0 : "n",
	1 : "e",
	2 : "w",
	3 : "s",
	4 : "u",
	5 : "d",
}

type Player struct {
	Name string
	ID int
	world map[int]*Room
	Location *Room
	Connection net.Conn
	Outputs chan OutputEvent
}

type InputEvent struct {
	player *Player
	Command string
	Login bool
}

type OutputEvent struct {
	mas string
}

var ALLROOM = make (map[int]*Room)

var ALLPLAYER = make (map[int]*Player)

var ConnToPlayer = make(map[net.Conn]*Player)

const listenAddress = ":3410"


// Part 1:
var allCommands = make(map[string]func(string, *Player))
// the MAKE bulid this type of data set in map and bulid the refrence

func commandLoop() error {
		player := ALLPLAYER[0]
		cmdwhere("", player)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			//log.Printf("The line was: %s", line)
	
			doCommand(line, player)
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("in main command loop: %v", err)
		}
	return nil
}

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
	addCommand("look", cmdLook)

	addCommand("north", cmdNorth)
	addCommand("east", cmdEast)
	addCommand("west", cmdWest)
	addCommand("south", cmdSouth)
	addCommand("up", cmdUp)
	addCommand("down", cmdDown)

	addCommand("where", cmdwhere)
	addCommand("recall", cmdRecall)

	addCommand("laugh", cmdLaugh)
	addCommand("tell", cmdTell)
}

func doCommand(cmd string, player *Player) error {
	words := strings.Fields(cmd)
	if len(words) == 0 {
		return nil
	}
	for i := 0; i < len(words) ; i++ {
		if f, exists := allCommands[strings.ToLower(words[i])]; exists {
			if strings.ToLower(words[i]) == "l" || strings.ToLower(words[i]) == "lo" || strings.ToLower(words[i]) == "loo" || strings.ToLower(words[i]) == "look"{
				cmdLook(words[i + 1], player)
				i++
			} else	{
				f(cmd, player)
			}
		}
	}
	return nil
}

func cmdLaugh(s string, player *Player) {
	if len(s) > 5 {
		log.Println("HAHAHAHAAHA! You laught about", s[5:])
	} else {
		fmt.Printf("\n")
		log.Println("HAHAHAHAAHA! You laught.")
	}
}

func cmdTell(s string, player *Player) {
	if len(s) > 4 {
		log.Println("You tell", s[4:])
	} else {
		fmt.Printf("\n")
		log.Println("You try to tell somebody about lsomething")
	}
}

func cmdLook(s string, player *Player) {
	if player.checkexit(string(s[0])) {
		look_next_loca := player.Location.Exits[Direction[string(s[0])]].Description
		fmt.Printf("\n")
		fmt.Printf(look_next_loca)
	} else {
		fmt.Printf("\n")
		fmt.Printf("There is nothing in that direcation.")
	}
}

func cmdwhere(s string, player *Player) {
	fmt.Printf(player.Location.Name)
	fmt.Printf("\n")
	fmt.Printf(player.Location.Description)
	fmt.Printf("\n")
	dir_list := []string{}
	for d := 0; d < 6; d++ {
		if player.Location.Exits[d].To != nil {
			dir_list = append(dir_list, ReDirection[d])
		}
	}
	fmt.Printf("[ Exits: %s ]\n", strings.Trim(fmt.Sprint(dir_list), "[]"))
}

func cmdRecall(s string, player *Player) {
	fmt.Printf("You take out a wishbone and pull apart. You see a light coming to you...\n")
	player.recall()
	cmdwhere(s, player)
}

func cmdNorth(s string, player *Player) {
	if player.checkexit("n") {
		next_loca := player.Location.Exits[Direction["n"]].To
		player = player.updatelocation(next_loca)
		cmdwhere(s,player)
	} else {
		fmt.Printf("You can't go that direction!\n")
	}
}

func cmdEast(s string, player *Player) {
	if player.checkexit("e") {
		player = player.updatelocation(player.Location.Exits[Direction["e"]].To)
		fmt.Printf("\n")
		cmdwhere(s,player)
	} else {
		fmt.Printf("You can't go that direction!\n")
	}
}

func cmdWest(s string, player *Player) {
	if player.checkexit("w") {
		player = player.updatelocation(player.Location.Exits[Direction["w"]].To)
		fmt.Printf("\n")
		cmdwhere(s,player)
	} else {
		fmt.Printf("You can't go that direction!\n")
	}
}

func cmdSouth(s string, player *Player) {
	if player.checkexit("s") {
		player = player.updatelocation(player.Location.Exits[Direction["s"]].To)
		fmt.Printf("\n")
		cmdwhere(s,player)
	} else {
		fmt.Printf("You can't go that direction!\n")
	}
}

func cmdUp(s string, player *Player) {
	if player.checkexit("u") {
		player = player.updatelocation(player.Location.Exits[Direction["u"]].To)
		fmt.Printf("\n")
		cmdwhere(s,player)
	} else {
		fmt.Printf("You can't go that direction!\n")
	}
}

func cmdDown(s string, player *Player) {
	if player.checkexit("d") {
		player = player.updatelocation(player.Location.Exits[Direction["d"]].To)
		fmt.Printf("\n")
		cmdwhere(s,player)
	} else {
		fmt.Printf("You can't go that direction!\n")
	}
}


// Part 2:

func loadzone (transaction *sql.Tx) (map[int]*Zone, error) {

	var id int
	//var zone_id int
	var name string
	//var description string

	var allzone = make(map[int]*Zone)

	rows, err := transaction.Query(`SELECT id, name FROM zones ORDER BY id`)
	if err != nil {
		return nil , err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}

		new_zone := Zone{id,name, nil}
		new_pointer := &new_zone
		allzone[id] = new_pointer
	}

	return allzone, nil
}

func loadroom (Zone map[int]*Zone  ,transaction *sql.Tx) (map[int]*Room, error) {

	var room_id int
	var room_name string
	var room_description string
	var allroom = make(map[int]*Room)
	var zone_id int

	rows, err := transaction.Query(`SELECT id, zone_id, name, description FROM rooms`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&room_id, &zone_id, &room_name, &room_description); err != nil {
			return nil, err
		}

		zone_pointer := Zone[zone_id] 
		new_room := Room{room_id, zone_pointer, room_name, room_description, [6]Exit{}}
		allroom[room_id] = &new_room
		Zone[zone_id].Rooms = []*Room{&new_room}

	}
	return allroom, nil
}

func loadExit (Room map[int]*Room, transaction *sql.Tx) (map[int]*Room, error) {
	var from_room  int
	var to_room int
	var direction string
	var description string

	rows, err := transaction.Query(`SELECT * FROM exits`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&from_room, &to_room, &direction, &description); err != nil {
			return nil, err
		}

		to_pointer := Room[to_room]
		new_exit := Exit{to_pointer, description}
		Room[from_room].Exits[Direction[direction]] = new_exit
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
	} else {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("Transaction failed while loading world data : %v", err)
		}
	}
	return nil
}

// Player methods: 

func initplayer(name string, world map[int]*Room , id int, conn net.Conn, Outputs chan OutputEvent) Player {
	player := Player{name, id, world, ALLROOM[3001], conn, Outputs}
	return player
}

func (player *Player) recall() {
	player.Location = ALLROOM[3001]
}

func (player *Player) updatelocation(new_room *Room) *Player{
	player.Location = new_room
	return player
}

func (player *Player) checkexit(dir string) bool{
	num_dir := Direction[dir]
	return player.Location.Exits[num_dir].To != nil 
}

func createplayer(conn net.Conn) error{
	var Name string
	fmt.Printf("Creating a player....\nType in your name?\n")
	fmt.Scanln(&Name)

	id := len(ALLPLAYER)
	Outputs := make(chan OutputEvent)

	// Create the target map
	world := make(map[int]*Room)

	// Copy from the original map to the target map
	for key, value := range ALLROOM {
		world[key] = value
	}

	player := initplayer(Name, world, id, conn, Outputs)

	ALLPLAYER[id] = &player

	ConnToPlayer[conn] = &player
	return nil
}

func setupWorld () error{
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

func setup() error{
	initCommands()
	err := setupWorld()
	if err != nil {
		return err
	} else {
		return nil
	}
}

func main () {
	err := setup()
	if err != nil {
		fmt.Printf("Setting up: %v", err)
	}

	inputChannel := make(chan InputEvent)
	go listenForConnections(inputChannel)
	mainLoop(inputChannel)
}

func mainLoop(inputChannel chan InputEvent) {
	for event := range inputChannel {
		
	}
}

func listenForConnections (inputChannel chan InputEvent) {
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
	createplayer(conn)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		event := InputEvent{player: ConnToPlayer[conn], Command: line, Login: true}
		inputChannel <- event
	} 
	if err := scanner.Err(); err != nil {
		log.Printf("connection error: %v", err)
	} else {
		log.Printf("connection closed normally\n")
	}
}



//func (p *Player) Printf (format string, a ...interface{}) {
//	mas := fmt.Sprintf(format, a...)
//	p.Outputs <- OutputEvent{Text: msg}
//}
