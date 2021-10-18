package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"fmt"
)

var allZones = make(map[int]*Zone)

type Zone struct {
    ID    int 
    Name  string
    //Rooms []*Room
}

func loadzoom () error {
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
	//var zone_id int
	var name string
	//var description string

	rows, err := db.Query(`SELECT id, name FROM zones`)
	if err != nil {
		log.Fatalf("scanning rows: %v", err)
	}

	for rows.Next() {
		if err := rows.Scan(&id, &name); err != nil {
			log.Fatalf("scanning row: %v", err)
		}

		new_zone := Zone{id, name}
		new_pointer := &new_zone
		allZones[id] = new_pointer
	}

	for key, value := range allZones {
		fmt.Println("id:", key, "pointer:", value)
	}

	return nil
}


func main () {
	err := loadzoom()
	if err != nil {
		log.Fatalf("main loading: %v", err)
	} else {
		fmt.Println("main loading: commit")
	}
}