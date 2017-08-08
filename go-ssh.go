package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"sync"
    "github.com/tmc/scp"
)

func Make_Config(id string, passwd string) *ssh.ClientConfig {
	config := &ssh.ClientConfig{
		User: id,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return config
}
func main() {
	opt := os.Args[1]
	addr := os.Args[2] + ":" + os.Args[3]
	id := os.Args[4]
	passwd := os.Args[5]
	var wg sync.WaitGroup
	config := Make_Config(id, passwd)

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	if opt == "-sh" {
		var cmd string
		for i := 6; i < len(os.Args); i++ {
			cmd = cmd + " " + os.Args[i]
		}
		wg.Add(1)
		go func(session *ssh.Session, cmd string) {
			defer wg.Done()
			defer session.Close()
			if err := session.Run(cmd); err != nil {
				log.Fatal("Failed to run: " + err.Error())
			}
		}(session, cmd)
	} else if opt == "-cp" {
		desPath := os.Args[len(os.Args)-1]
		for i := 6; i < len(os.Args)-1; i++ {
			wg.Add(1)
			go func(filePath, desPath string, client *ssh.Client) {
				session, err := client.NewSession()
				if err != nil {
					log.Fatal("Failed to create session: ", err)
				}
				defer wg.Done()
				err = scp.CopyPath(filePath, desPath, session)
				if err != nil {
					log.Fatal("Failed to scp: " + err.Error())
				}
			}(os.Args[i], desPath, client)
		}
	} else {
		fmt.Println("Usage : <-sh|-cp> ADDR ID PASSWD <CMD|SOURCE FILE> <|DESTINATION FILE>")
	}
	wg.Wait()
	fmt.Println("Complete")
}
