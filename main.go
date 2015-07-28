package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage")
	}

	path, err := exec.LookPath("./" + os.Args[1])
	if err != nil {
		path, err = exec.LookPath(os.Args[1])
		if err != nil {
			os.Exit(127)
		}
	}

	sigs := make(chan os.Signal, 1) // There is a go routine so I think capacity 1 is enough
	signal.Notify(sigs, syscall.SIGHUP)

	cmd := exec.Command(path)
	cmd.Args = os.Args[2:]
	// WTF is S_IRUSR | S_IWUSR, hoping its 0600
	logFile, err := os.OpenFile("nohup.out", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal("TODO: create it in $HOME")
	}
	cmd.Stdout = logFile
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	go handleSigs(sigs)
	fmt.Println("Pid of the started process: ", cmd.Process.Pid)
	cmd.Wait()
	sendEmail()
}

func handleSigs(sigs chan os.Signal) {
	for {
		<-sigs
	}
}

func sendEmail() {

}
