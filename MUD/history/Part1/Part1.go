package main

import (
	"fmt"
	"log"
	"bufio"
	"os"
	"strings"
)

var allCommands = make(map[string]func(string))
// the MAKE bulid this type of data set in map and bulid the refrence

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Println("hello, world!")
	//fmt.Println("Hello, world!")
	if err := commandLoop(); err != nil {
		log.Fatalf("%v", err)
	} 
}

func commandLoop() error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("The line was: %s", line)

		initCommands()
		doCommand(line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("in main command loop: %v", err)
	}

	return nil
}

func addCommand(cmd string, f func(string)) {
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
	addCommand("smile", cmdSmile)
	addCommand("south", cmdSouth)
	addCommand("look", cmdLook)
	addCommand("sigh", cmdSigh)
	addCommand("laugh", cmdLaugh)
	addCommand("say", cmdSay)
	addCommand("tell", cmdTell)
}

func doCommand(cmd string) error {
	words := strings.Fields(cmd)
	if len(words) == 0 {
		return nil
	}
	if f, exists := allCommands[strings.ToLower(words[0])]; exists {
		f(cmd)
	}
	return nil
}

func cmdSmile(a string) {
	if len(a) > 5 {
		log.Println("You smile", a[5:])
	} else {
		log.Println("You smile.")
	}
}

func cmdSouth(a string) {
	if len(a) > 5 {
		log.Println("You are facing the south and", a[5:])
	} else {
		log.Println("You are facing the south.")
	}
}

func cmdLook(a string) {
	if len(a) > 4 {
		log.Println("You take a look at", a[4:])
	} else {
		log.Println("You look around.")
	}
}

func cmdSigh(a string) {
	if len(a) >= 4 {
		log.Println("*Sigh*", a[4:])
	} else {
		log.Println("*Sigh*")
	}
}

func cmdLaugh(a string) {
	if len(a) > 5 {
		log.Println("HAHAHAHAAHA! You laught about", a[5:])
	} else {
		log.Println("HAHAHAHAAHA! You laught.")
	}
}

func cmdSay(a string) {
	if len(a) > 3 {
		log.Println("You say:", a[3:])
	} else {
		log.Println("You try to say somthing.")
	}

}

func cmdTell(a string) {
	if len(a) > 4 {
		log.Println("You tell", a[4:])
	} else {
		log.Println("You try to tell somebody about lsomething")
	}
}