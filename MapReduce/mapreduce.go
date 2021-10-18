package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"mapreduce"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	defaultHost = "localhost"
	defaultPort = "3410"
)

type Master struct {
	Address           string
	TaskList          *Task
	M, R, MS, RS      int   // MS: map sent RS: reduce sent
	MF, RF            []int // MS: map finished RS: reduce finished
	FinishTaskAddList *FinishList
}

type FinishList struct {
	Map    []string
	Reduce []string
}

type Task struct {
	Maptasklist    []*mapreduce.MapTask
	Reducetasklist []*mapreduce.ReduceTask
}

type In struct {
	Id      string   // node ID
	Type    string   // M R and nil
	Number  int      // task number
	FileAdd []string // finish task address
}
type Msg struct {
	Header string // S: sleep F: finish M: map R:reduce
	M      *mapreduce.MapTask
	R      *mapreduce.ReduceTask
}

type handler func(*Master)
type Server chan<- handler

var done = make(chan struct{})

type Nothing struct{}

func clean(tempdir string) {
	err := os.RemoveAll(tempdir)
	if err != nil {
		log.Fatal(err)
	}
}

func printUsage() {
	log.Printf("Usage: %s [-master or (default as -worker)] [address]", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	log.Printf("current directory %s\n", root)
	masterAdd := net.JoinHostPort(defaultHost, defaultPort)
	workerAdd := net.JoinHostPort(defaultHost, defaultPort)
	masteDir := filepath.Join(root, "/test/part3/master")
	workerDir := filepath.Join(root, "/test/part3/worker")

	var isMaster bool
	var isWorker bool
	port := defaultPort
	flag.BoolVar(&isMaster, "master", false, "start as master")
	flag.BoolVar(&isWorker, "worker", true, "start as worker")
	flag.Parse()

	switch flag.NArg() {
	case 0:
		break
	case 1:
		port = flag.Arg(0)
		workerAdd = net.JoinHostPort(defaultHost, port)
	default:
		printUsage()
	}

	if isMaster {
		log.Print("start as master\n")
	loop:
		for {
			select {
			case <-done:
				break loop
			default:
				shell(masterAdd, masteDir)
			}
		}
	} else {
		if port != defaultPort {
			log.Print("start as worker\n")
			worker(masterAdd, workerDir, workerAdd)
		} else {
			log.Print("please specify the address\n")
		}
	}
	os.RemoveAll(workerDir)
	fmt.Print("program is closing\n")
}

func worker(masterAdd string, dir string, add string) {
	_, port, err := net.SplitHostPort(add)
	if err != nil {
		log.Fatalf("worker net.SplitHostPort %v", err)
	}
	log.Printf("worker id: %s address: %s dir: %s\n", port, add, dir)

	go func() {
		http.Handle("/data/", http.StripPrefix("/data", http.FileServer(http.Dir(dir))))
		if err := http.ListenAndServe(add, nil); err != nil {
			log.Printf("Error in HTTP server for %s: %v", add, err)
		}
	}()
	log.Print("worker server start\n")
	var c Client
	var in In
	in.Id = port
	in.Type = ""
	for {
		var responce *Msg
		log.Print("send message to master\n")
		if err = call(masterAdd, "Server.Task", &in, &responce); err != nil {
			log.Fatal(err)
		}
		if responce.Header != "" {
			log.Print("recived message from master\n")
		}
		if responce.Header == "S" {
			log.Print("sleeping...")
			time.Sleep(2 * time.Second)
		} else if responce.Header == "F" {
			log.Print("finish all tasks!\n")
			time.Sleep(15 * time.Second)
			break
		} else if responce.Header == "M" {
			log.Printf("working on Map task %d\n", responce.M.N)
			err := responce.M.Process(dir, c)
			if err != nil {
				log.Fatalf("responce.M.Process %v", err)
			}
			log.Printf("finish Map task %d\n", responce.M.N)
			var file []string
			for i := 0; i < responce.M.R; i++ {
				file = append(file, fmt.Sprintf("http://%s/data/%s", add, fmt.Sprintf("map_%d_output_%d.db", responce.M.N, i)))
			}
			log.Printf("file address %s\n", file)
			in.Type = "M"
			in.Number = responce.M.N
			in.FileAdd = file
		} else if responce.Header == "R" {
			log.Printf("working on Reduce task %d\n", responce.R.N)
			err := responce.R.Process(dir, c)
			if err != nil {
				log.Fatalf("responce.R.Process %v", err)
			}
			log.Printf("finish Reduce task %d\n", responce.R.N)
			var file []string
			file = append(file, fmt.Sprintf("http://%s/data/%s", add, fmt.Sprintf("reduce_%d_output.db", responce.R.N)))
			log.Printf("file address %s\n", file)
			in.Type = "R"
			in.Number = responce.R.N
			in.FileAdd = file
		}
		log.Print("-----------\n")
	}
	// clean up
	// os.RemoveAll(dir)
}

func shell(add string, dir string) {
	log.Printf("Starting interactive shell")
	help()
	master := Master{M: 9, R: 3, MF: make([]int, 9), RF: make([]int, 3), Address: add, MS: 0, RS: 0, TaskList: &Task{[]*mapreduce.MapTask{}, []*mapreduce.ReduceTask{}}, FinishTaskAddList: &FinishList{[]string{}, []string{}}}
	log.Printf("Map tasks: %d\tReduce tasks: %d\n", master.M, master.R)
	log.Printf("temporary directory: %s\n", dir)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "port":
			if len(parts) == 2 {
				port := parts[1]
				master.Address = net.JoinHostPort(defaultHost, port)
			}
			fmt.Printf("The address for listening is : %s\n", master.Address)
			continue
		case "M":
			if len(parts) == 2 {
				m, err := strconv.Atoi(parts[1])
				master.M = m
				master.MF = make([]int, master.M)
				if err == nil {
					log.Printf("Map tasks: %d\tReduce tasks: %d\n", master.M, master.R)
					continue
				}
			}
			log.Printf("can't not identify, please try again\n")
			continue
		case "R":
			if len(parts) == 2 {
				r, err := strconv.Atoi(parts[1])
				master.R = r
				master.RF = make([]int, master.R)
				if err == nil {
					log.Printf("Map tasks: %d\tReduce tasks: %d\n", master.M, master.R)
					continue
				}
			}
			log.Printf("can't not identify, please try again\n")
			continue
		case "target":
			if len(parts) == 2 {
				filename := parts[1]
				if _, err := os.Stat(filename); os.IsNotExist(err) {
					log.Println("file does not exist")
					current, err := os.Getwd()
					if err != nil {
						log.Println(err)
					}
					log.Printf("file name %s \t current dir %s\n", filename, current)
				} else {
					log.Printf("%s is set\n", filename)

					log.Printf("spliting the target file %s\n", filename)
					sourefiles, err := mapreduce.SplitDatabase(filename, dir, "map_%d_source.db", master.M)
					if err != nil {
						log.Fatalf("master spliting source file %v", err)
					}
					for i := range sourefiles {
						_, sourefiles[i] = filepath.Split(sourefiles[i])
					}
					log.Printf("finished spliting\ncreated source files %s\n", sourefiles)
					for i := range sourefiles {
						sourefiles[i] = fmt.Sprintf("http://%s/data/%s", net.JoinHostPort(defaultHost, defaultPort), sourefiles[i])
					}
					// Generate the full set of map tasks and reduce tasks
					var maptasklist []*mapreduce.MapTask
					var reducetasklist []*mapreduce.ReduceTask
					for i := 0; i < master.M; i++ {
						task := mapreduce.MapTask{M: master.M, R: master.R, N: i, SourceHost: sourefiles[i]}
						maptasklist = append(maptasklist, &task)
					}
					// SourceHosts stil empty, should be starded filling up after A Map task done
					for i := 0; i < master.R; i++ {
						task := mapreduce.ReduceTask{M: master.M, R: master.R, N: i}
						reducetasklist = append(reducetasklist, &task)
					}
					master.TaskList.Maptasklist = maptasklist
					master.TaskList.Reducetasklist = reducetasklist
					log.Printf("created %d Map tasks and %d Reduce tasks\n", master.M, master.R)

					// Create and start an RPC server
					go func() {
						http.Handle("/data/", http.StripPrefix("/data", http.FileServer(http.Dir(dir))))
						actor := startActor(&master)
						rpc.Register(actor)
						rpc.HandleHTTP()
						if err := http.ListenAndServe(":3410", nil); err != nil {
							log.Printf("Error in HTTP server for %s: %v", ":3410", err)
						}
					}()
					log.Print("started PRC server\n")
					// Wait until all jobs are complete
					for {
						wait := false
						time.Sleep(3 * time.Second)
						if len(checkfinish(master.RF)) == master.R {
							wait = true
						}
						if wait {
							break
						}
					}

					// Gather the reduce outputs and join them into a single output file.
					finaloutputDB, err := mapreduce.MergeDatabases(master.FinishTaskAddList.Reduce, filepath.Join(dir, "mapreduce.db"), filepath.Join(dir, "temp.db"))
					if err != nil {
						log.Fatalf("Gather all the reduce outputs %v", err)
					}
					log.Print("final marge done\n")
					defer finaloutputDB.Close()

					// finished all tasks
					// clean(tempdir)
					close(done)
					return
				}
				continue
			} else {
				log.Printf("can't not identify, please try again\n")
				continue
			}
		case "quit":
			clean(dir)
			close(done)
			return
		case "dump":
			log.Printf("adddress: %s", master.Address)
			log.Printf("Map tasks: %d\tReduce tasks: %d\n", master.M, master.R)
			log.Printf("temporary directory: %s", dir)
		case "help":
			fallthrough
		default:
			help()
		}
	}
}

func startActor(master *Master) Server {
	ch := make(chan handler)
	state := master
	go func() {
		for f := range ch {
			f(state)
		}
	}()
	return ch
}

func (s Server) Task(in *In, out *Msg) error {
	log.Printf("rescived request from %s\n", in.Id)
	finished := make(chan struct{})
	s <- func(m *Master) {
		// read in message
		if in.Type == "M" && m.MF[in.Number] != 1 {
			log.Printf("finished Map task %d\n", in.Number)
			log.Printf("%d files contained in the input\n", len(in.FileAdd))
			for i := range in.FileAdd {
				m.FinishTaskAddList.Map = append(m.FinishTaskAddList.Map, in.FileAdd[i])
			}
			m.MF[in.Number] = 1
			for i := 0; i < m.R; i++ {
				m.TaskList.Reducetasklist[i].SourceHosts = append(m.TaskList.Reducetasklist[i].SourceHosts, in.FileAdd[i])
			}
		} else if in.Type == "R" && m.RF[in.Number] != 1 {
			log.Printf("finished Reduce task %d\n", in.Number)
			log.Printf("%d files contained in the input\n", len(in.FileAdd))
			for i := range in.FileAdd {
				m.FinishTaskAddList.Reduce = append(m.FinishTaskAddList.Reduce, in.FileAdd[i])
			}
			m.RF[in.Number] = 1
		}

		// send out maptask
		if m.MS < m.M {
			log.Printf("send Map task %d to %s\n", m.MS, in.Id)
			*out = Msg{Header: "M", M: m.TaskList.Maptasklist[m.MS]}
			m.MS++
		} else if len(checkfinish(m.MF)) < m.M {
			log.Printf("tell %s to sleep because Maps task not finish yet\n", in.Id)
			*out = Msg{Header: "S"}
		}

		// finished all the maps start sending reduce task
		if len(checkfinish(m.MF)) == m.M {
			if m.RS < m.R {
				log.Printf("send Reduce task %d to %s\n", m.RS, in.Id)
				*out = Msg{Header: "R", R: m.TaskList.Reducetasklist[m.RS]}
				m.RS++
			} else if len(checkfinish(m.RF)) < m.R {
				log.Printf("tell %s to sleep because Reduce tasks not finish yet\n", in.Id)
				*out = Msg{Header: "S"}
			} else if len(checkfinish(m.RF)) == m.R {
				log.Printf("tell %s all work are finished, start shut dowm process\n", in.Id)
				*out = Msg{Header: "F"}
			}
		}

		log.Printf("sent responce to node %s\n", in.Id)
		log.Printf("Tasks check:\nMap tasks(%d): sent %d finished %d\nReduce tasks(%d): sent %d finished %d\n", m.M, m.MS, len(checkfinish((m.MF))), m.R, m.RS, len(checkfinish(m.RF)))
		log.Print("---------------\n")
		finished <- struct{}{}
	}
	<-finished
	return nil
}

func checkfinish(lst []int) []int {
	var check []int
	for i := 0; i < len(lst); i++ {
		if lst[i] == 1 {
			check = append(check, lst[i])
		}
	}
	return check
}

func help() {
	fmt.Print("Only those commands are supported\n")
	fmt.Printf("\t%-20s\t change the port should be listened\n", "port <port number>")
	fmt.Printf("\t%-20s\t file will be run MapReduce\n", "target <filename>")
	fmt.Printf("\t%-20s\t change the total number of Map tasks\n", "M <number>")
	fmt.Printf("\t%-20s\t change the total number of Reduce tasks\n", "R <number>")
	fmt.Printf("\t%-20s\t show master information\n", "dump")
	fmt.Printf("\t%-20s\t exit the program\n", "quit")
	fmt.Printf("\t%-20s\t print out help message\n", "help")
}

func call(serverAddress string, method string, request interface{}, responce interface{}) error {
	client, err := rpc.DialHTTP("tcp", serverAddress)
	if err != nil {
		log.Fatalf("rpc.DialHTTP: %v", err)
		return err
	}
	defer client.Close()

	if err = client.Call(method, request, responce); err != nil {
		return err
	}

	return nil
}

// Map and Reduce functions for a basic wordcount client

type Client struct{}

func (c Client) Map(key, value string, output chan<- mapreduce.Pair) error {
	defer close(output)
	lst := strings.Fields(value)
	for _, elt := range lst {
		word := strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				return unicode.ToLower(r)
			}
			return -1
		}, elt)
		if len(word) > 0 {
			output <- mapreduce.Pair{Key: word, Value: "1"}
		}
	}
	return nil
}

func (c Client) Reduce(key string, values <-chan string, output chan<- mapreduce.Pair) error {
	defer close(output)
	count := 0
	for v := range values {
		i, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		count += i
	}
	p := mapreduce.Pair{Key: key, Value: strconv.Itoa(count)}
	output <- p
	return nil
}
