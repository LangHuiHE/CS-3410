package mapreduce

import (
	"database/sql"
	"fmt"
	"hash/fnv"
	"log"
	"path/filepath"
)

type MapTask struct {
	M, R       int    // total number of map and reduce tasks
	N          int    // map task number, 0-based
	SourceHost string // address of host with map input file
}

type ReduceTask struct {
	M, R        int      // total number of map and reduce tasks
	N           int      // reduce task number, 0-based
	SourceHosts []string // addresses of map workers
}

type Pair struct {
	Key   string
	Value string
}

type Interface interface {
	Map(key, value string, output chan<- Pair) error
	Reduce(key string, values <-chan string, output chan<- Pair) error
}

func mapSourceFile(m int) string       { return fmt.Sprintf("map_%d_source.db", m) }
func mapInputFile(m int) string        { return fmt.Sprintf("map_%d_input.db", m) }
func mapOutputFile(m, r int) string    { return fmt.Sprintf("map_%d_output_%d.db", m, r) }
func reduceInputFile(r int) string     { return fmt.Sprintf("reduce_%d_input.db", r) }
func reduceOutputFile(r int) string    { return fmt.Sprintf("reduce_%d_output.db", r) }
func reducePartialFile(r int) string   { return fmt.Sprintf("reduce_%d_partial.db", r) }
func reduceTempFile(r int) string      { return fmt.Sprintf("reduce_%d_temp.db", r) }
func makeURL(host, file string) string { return fmt.Sprintf("http://%s/data/%s", host, file) }

// MapTask.Process
func (task *MapTask) Process(tempdir string, client Interface) error {
	err := download(task.SourceHost, filepath.Join(tempdir, mapInputFile(task.N)))
	if err != nil {
		return err
	}
	path := filepath.Join(tempdir, mapInputFile(task.N))
	inputFile, err := openDatabase(path)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	stmtList := make([]*sql.Stmt, task.R)
	for i := 0; i < task.R; i++ {
		fileName := filepath.Join(tempdir, mapOutputFile(task.N, i))
		outputFile, err := createDatabase(fileName)
		if err != nil {
			return err
		}
		defer outputFile.Close()

		stmt, err := outputFile.Prepare("insert into pairs (key, value) values (?, ?)")
		if err != nil {
			return err
		}

		stmtList[i] = stmt
	}

	tx, err := inputFile.Begin()
	if err != nil {
		log.Fatal(err)
	}
	rows, err := tx.Query("select key, value from pairs")
	if err != nil {
		log.Fatalf("MapTask.Process input.Query: %s %v\n", path, err)
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var outputs = make(chan Pair)
		go func() error {
			for pair := range outputs {

				hash := fnv.New32() // from the stdlib package hash/fnv
				hash.Write([]byte(pair.Key))
				r := int(hash.Sum32() % uint32(task.R))
				if _, err := stmtList[r].Exec(pair.Key, pair.Value); err != nil {
					return err
				}
				/*
					_, err = outputFile.Exec(`insert into pairs (key, value) values (?, ?)`, pair.Key, pair.Value)
					if err != nil {
						log.Fatalf("MapTask.Process db.Exec: %v\n", err)
						return err
					}
				*/
				if err = rows.Err(); err != nil {
					return err
				}
			}
			return nil
		}()

		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			log.Fatalf("MapTask.Process rows.Scan: %v\n", err)
			return err
		}
		if err = client.Map(key, value, outputs); err != nil {
			return err
		}
	}

	for i := range stmtList {
		stmtList[i].Close()
	}
	return nil
}

// ReduceTask.Process
func (task *ReduceTask) Process(tempdir string, client Interface) error {
	inputDB, err := MergeDatabases(task.SourceHosts, filepath.Join(tempdir, reduceInputFile(task.N)), filepath.Join(tempdir, reducePartialFile(task.N)))
	if err != nil {
		return err
	}
	log.Print("finish merging input file\n")
	path := filepath.Join(tempdir, reduceOutputFile(task.N))
	log.Print(path)
	outputDB, err := createDatabase(path)
	if err != nil {
		log.Print(path)
		return err
	}

	rows, err := inputDB.Query("select key, value from pairs order by key, value")
	if err != nil {
		return err
	}
	defer rows.Close()

	previous := ""
	input := make(chan string)
	output := make(chan Pair)
	var key, value string
	for rows.Next() {
		if err := rows.Scan(&key, &value); err != nil {
			log.Fatalf("ReduceTask.Process rows.Scan: %v\n", err)
			return err
		}
		if previous != key {
			if previous != "" {
				close(input)

				for i := range output {
					_, err := outputDB.Exec(`insert into pairs (key, value) values (?, ?)`, i.Key, i.Value)
					if err != nil {
						log.Fatalf("middle ReduceTask.Process db.Exec: %v\n", err)
					}
				}

				input = make(chan string)
				output = make(chan Pair)
			}

			previous = key
			go func() {
				err := client.Reduce(key, input, output)
				if err != nil {
					log.Fatalf("row.Next client.Reduce %v\n", err)
				}
			}()
			input <- value
		} else {
			input <- value
		}
	}
	// close the last call
	close(input)

	for i := range output {
		_, err := outputDB.Exec(`insert into pairs (key, value) values (?, ?)`, i.Key, i.Value)
		if err != nil {
			log.Fatalf("last ReduceTask.Process db.Exec: %v\n", err)
		}
	}

	defer inputDB.Close()
	defer outputDB.Close()
	return nil
}
