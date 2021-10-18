package main

import (
	"database/sql"
	"fmt"
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

// OutputEvent struct
type OutputEvent struct {
	Mag string
}

// ALLROOM all the room
var ALLROOM = make(map[int]*Room)

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

// SetupWorld setup
func SetupWorld() error {
	// Part 2:
	// the path to the database--this could be an absolute path
	//path := "world.db"
	path := "/Users/Aaron/Desktop/CS 3410/MUD/world.db"
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
