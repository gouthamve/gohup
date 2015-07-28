package main

import (
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

	args := os.Args[2:]
	// WTF is S_IRUSR | S_IWUSR, hoping its 0600
	logFile, err := os.OpenFile("nohup.out", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal("TODO: create it in $HOME")
	}

	syscall.Dup2(int(logFile.Fd()), 1)
	syscall.Dup2(int(logFile.Fd()), 2)

	env := os.Environ()
	execErr := syscall.Exec(path, args, env)
	if execErr != nil {
		panic(execErr)
	}
	go handleSigs(sigs)
}

func handleSigs(sigs chan os.Signal) {
	for {
		<-sigs
	}
}
