package mapreduce

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func openDatabase(path string) (*sql.DB, error) {
	options :=
		"?" + "_busy_timeout=10000" +
			"&" + "_case_sensitive_like=OFF" +
			"&" + "_foreign_keys=ON" +
			"&" + "_journal_mode=OFF" +
			"&" + "_locking_mode=NORMAL" +
			"&" + "mode=rw" +
			"&" + "_synchronous=OFF"
	db, err := sql.Open("sqlite3", path+options)
	if err != nil {
		// handle the error here
		return nil, err
	}
	return db, nil
}

func createDatabase(path string) (*sql.DB, error) {
	// delete exist database
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		os.Remove(path)
	}

	// create a dababase
	db, err := openDatabase(path)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("create table pairs (key text, value text)")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func opentransaction(database *sql.DB) (*sql.Tx, error) {
	tx, err := database.Begin()
	if err != nil {
		log.Fatalf("opentransaction database.Begin: %v\n", err)
		return nil, err
	}
	return tx, nil
}

func SplitDatabase(source, outputDir, outputPattern string, m int) ([]string, error) {
	// load the data
	var count int
	var key, value string
	var list [][]string
	var i int

	db, err := openDatabase(source)
	if err != nil {
		log.Print(source)
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("select key, value from pairs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		// read the data using rows.Scan
		// process the result
		if err := rows.Scan(&key, &value); err != nil {
			log.Fatalf("rows.Scan: %v\n", err)
			return nil, err
		}
		list = append(list, []string{})
		list[i] = append(list[i], key)
		list[i] = append(list[i], value)
		i++
		count++
	}
	if err = rows.Err(); err != nil {
		// handle the error
		return nil, err
	}
	if count < m {
		countErr := errors.New("read row count is smaller than m\n")
		return nil, countErr
	}

	// create the output files
	var outputDBs []string
	count = 0

	for i := 0; i < m; i++ {
		newPath := filepath.Join(outputDir, fmt.Sprintf(outputPattern, i))
		if _, err := createDatabase(newPath); err != nil {
			log.Printf("splitDatabase create database: %v\n", err)
			return nil, err
		}
		outputDBs = append(outputDBs, newPath)

		db, err := openDatabase(newPath) // open the output database files
		if err != nil {
			log.Fatalf("splitDatabase open database: %v\n", err)
			return nil, err
		}
		defer db.Close()
		// insert into the output database files
		listIndex := i
		for listIndex < len(list) {
			key = list[listIndex][0]
			value = list[listIndex][1]

			_, err = db.Exec(`insert into pairs (key, value) values (?, ?)`, key, value)
			if err != nil {
				log.Fatalf("splitDatabase db.Exec: %v\n", err)
				return nil, err
			}
			defer db.Close()
			count++
			listIndex += m
		}
		db.Close()

		if count < m {
			countErr := errors.New("insert row count is smaller than m\n")
			return nil, countErr
		}
	}
	return outputDBs, nil
}

func MergeDatabases(urls []string, path string, temp string) (*sql.DB, error) {
	// Create a new output database
	db, err := createDatabase(path)
	if err != nil {
		return nil, err
	}
	if db, err = openDatabase(path); err != nil {
		return nil, err
	}

	for i := range urls {
		if err := download(urls[i], temp); err != nil {
			log.Fatalf("download: %v\n", err)
			return nil, err
		}
		if err = gatherInto(db, temp); err != nil {
			return nil, err
		}
	}

	return db, nil
}

// out.db, pathname(input file)
func gatherInto(db *sql.DB, path string) error {
	_, err := db.Exec("attach ? as merge;insert into pairs select * from merge.pairs;detach merge;", path)
	if err != nil {
		return err
	}
	os.Remove(path)
	return nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func download(url, path string) error {
	// Get the data
	resp, err := http.Get(url)
	log.Printf("download from url %s\n", url)
	if err != nil {
		log.Print(url)
		log.Fatalf("download Get: %v\n", err)
		return err
	}
	defer resp.Body.Close()
	// Create the file
	out, err := os.Create(path)
	if err != nil {
		log.Fatalf("download os.Create: %v\n", err)
		return err
	}
	defer out.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("bad status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatalf("download io.Copy: %v\n", err)
		return err
	}
	return nil
}
