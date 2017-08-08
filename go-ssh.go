package main

import (
	"encoding/json"
	"fmt"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"sync"
    "io/ioutil"
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

type Config struct {
	Addr   string
	Role   string
	Port   string
	Passwd string
}

func main() {
	opt := os.Args[1]
	file, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		log.Fatal("Fail to read JSON: ", err)
	}
	var data []Config
	json.Unmarshal(file, &data)
	var wg sync.WaitGroup
	for i := 0; i < len(data); i++ {
		wg.Add(1)
		go func(ip, port, passwd string) {
			defer wg.Done()
			addr := ip + ":" + port
			config := Make_Config("root", passwd)

			client, err := ssh.Dial("tcp", addr, config)
			if err != nil {
				log.Fatal("Failed to dial: ", err)
			}
			if opt == "-sh" {
				session, err := client.NewSession()
				defer session.Close()
				if err != nil {
					log.Fatal("Failed to create session: ", err)
				}
				cmd := os.Args[3]
				if err := session.Run(cmd); err != nil {
					log.Fatal("Failed to run: " + err.Error())
				}
			} else if opt == "-cp" {
				desPath := os.Args[len(os.Args)-1]
				var cp_wg sync.WaitGroup
				for j := 3; j < len(os.Args)-1; j++ {
					cp_wg.Add(1)
					go func(filePath, desPath string, client *ssh.Client) {
						session, err := client.NewSession()
                            defer session.Close()
						if err != nil {
							log.Fatal("Failed to create session: ", err)
						}
						defer cp_wg.Done()
						err = scp.CopyPath(filePath, desPath, session)
						if err != nil {
							log.Fatal("Failed to scp: " + err.Error())
						}
					}(os.Args[j], desPath, client)
				}
				cp_wg.Wait()
			}
		}(data[i].Addr, data[i].Port, data[i].Passwd)
	}
	wg.Wait()
	fmt.Println("Complete")
}
