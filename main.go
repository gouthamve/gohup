package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strconv"
	"syscall"

	"github.com/Unknwon/goconfig"
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
	sendEmail(cmd)
}

func handleSigs(sigs chan os.Signal) {
	for {
		<-sigs
	}
}

func sendEmail(cmd *exec.Cmd) {
	curUser, err := user.Current()
	if err != nil {
		return
	}

	cfg, err := goconfig.LoadConfigFile(curUser.HomeDir + "/.nohup")
	if err != nil {
		return
	}

	fromEmail, err := cfg.GetValue("", "fromemail")
	if err != nil {
		return
	}

	toEmail, err := cfg.GetValue("", "toemail")
	if err != nil {
		return
	}

	password, err := cfg.GetValue("", "password")
	if err != nil {
		return
	}

	type EmailConfig struct {
		Username string
		Password string
		Host     string
		Port     int
	}

	// authentication configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := 587
	smtpPass := password
	smtpUser := fromEmail

	emailConf := &EmailConfig{smtpUser, smtpPass, smtpHost, smtpPort}

	emailauth := smtp.PlainAuth("", emailConf.Username, emailConf.Password, emailConf.Host)

	sender := fromEmail // change here

	receivers := []string{
		toEmail,
	}

	message := []byte("Your Process, " + cmd.Path + "has stopped") // your message

	// send out the email
	err = smtp.SendMail(smtpHost+":"+strconv.Itoa(emailConf.Port), //convert port number from int to string
		emailauth,
		sender,
		receivers,
		message,
	)

	if err != nil {
		fmt.Println(err)
	}
}
