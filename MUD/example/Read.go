package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"fmt"
)


func loadroom () {
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
		log.Fatalf("opening database: %v", err)
	}  

	defer db.Close()

	var id int
	var zone_id int
	var name string
	var description string

	err = db.QueryRow(`SELECT id, zone_id, name, description FROM rooms WHERE id = 3001`).Scan(
		&id, &zone_id, &name, &description)
	if err != nil {
		log.Fatalf("scanning room: %v", err)
	}


	fmt.Printf("id:%d zoneID:%d name:%s\ndescription:%s\n", id, zone_id, name, description)
}


func main () {
	loadroom()
}