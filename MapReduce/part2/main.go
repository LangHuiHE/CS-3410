package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	//runtime.GOMAXPROCS(1)

	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	r := 3

	address := "localhost:8080"
	/*
		m := 9
		tempdir := filepath.Join(root, "test/part2/map")
			source := filepath.Join(root, "test/split/in/austen.db")
			outputPattern := "map_%d_input.db"

			go func() {
				http.Handle("/data/", http.StripPrefix("/data", http.FileServer(http.Dir(tempdir))))
				if err := http.ListenAndServe(address, nil); err != nil {
					log.Printf("Error in HTTP server for %s: %v", address, err)
				}
			}()

			_, err = splitDatabase(source, tempdir, outputPattern, m)
			if err != nil {
				log.Fatal(err)
			}

			//fmt.Printf("the lst from split : %s\n", lst)

				for n := 0; n < m; n++ {
					m := MapTask{M: m, R: r, N: n, SourceHost: makeURL("localhost:8080/", "austen.db")}
					if err = m.Process(tempdir, Client{}); err != nil {
						log.Fatal(err)
					}
				}

				go func() {
					http.Handle("/data/", http.StripPrefix("/data", http.FileServer(http.Dir(filepath.Join(root, "test/part2/map")))))
					if err := http.ListenAndServe(address, nil); err != nil {
						log.Printf("Error in HTTP server for %s: %v", address, err)
					}
				}()

				tempdir = filepath.Join(root, "test/part2/reduce")
				for i := 0; i < r; i++ {
					var lst []string
					for j := 0; j < m; j++ {
						file := mapOutputFile(j, i)
						lst = append(lst, makeURL(address, file))
					}
					m := ReduceTask{M: m, R: r, N: i, SourceHosts: lst}
					if err = m.Process(tempdir, Client{}); err != nil {
						log.Fatal(err)
					}
				}

	*/
	// final merger
	var urls []string
	for i := 0; i < r; i++ {

		file := reduceOutputFile(i)
		urls = append(urls, makeURL(address, file))
	}
	log.Print(urls)
	after := filepath.Join(root, "test/part2/reduce/reduce_out.db")
	go func() {
		http.Handle("/data/", http.StripPrefix("/data", http.FileServer(http.Dir(filepath.Join(root, "test/part2/reduce")))))
		if err := http.ListenAndServe(address, nil); err != nil {
			log.Printf("Error in HTTP server for %s: %v", address, err)
		}
	}()
	_, err = mergeDatabases(urls, after, filepath.Join(root, "test/part2/map"))
	if err != nil {
		log.Fatal(err)
	} else {
		log.Print("\ndone")
	}

	//defer os.RemoveAll(tempdir)
}
