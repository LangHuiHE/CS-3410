package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

func main() {
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	source := filepath.Join(root, "test/split/in/austen.db")
	outputDir := filepath.Join(root, "/test/split/out")
	outputPattern := "austen_%d.db"
	m := 2
	filePaths, err := splitDatabase(source, outputDir, outputPattern, m)
	if err != nil {
		log.Fatal(err)
	}

	var urls []string
	for i := range filePaths {
		u, err := url.Parse("http://localhost:8080/data/")
		if err != nil {
			log.Fatal(err)
		}
		_, fileName := filepath.Split(filePaths[i])
		u.Path = path.Join(u.Path, fileName)
		urls = append(urls, u.String())
	}

	address := "localhost:8080"
	tempdir := filepath.Join(root, "/test/split/out")
	after := filepath.Join(root, "/test/merge/out/afterMerge.db")
	temp := filepath.Join(root, "/test/merge/temp")
	go func() {
		http.Handle("/data/", http.StripPrefix("/data", http.FileServer(http.Dir(tempdir))))
		if err := http.ListenAndServe(address, nil); err != nil {
			log.Printf("Error in HTTP server for %s: %v", address, err)
		}
	}()

	_, err = mergeDatabases(urls, after, temp)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Print("\ndone")
	}

}
