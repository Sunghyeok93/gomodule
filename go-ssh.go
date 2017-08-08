package main

import (
	"encoding/json"
	"fmt"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"sync"
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

type Opt struct {
	opt_A int
	opt_M int
	opt_C int
	opt_S int
}

func check_opt() Opt {
	var opts Opt
	opts.opt_A = 0
	opts.opt_M = 0
	opts.opt_S = 0
	opts.opt_C = 0
	for i := 0; i < len(os.Args); i++ {
		if os.Args[i] == "-a" {
			opts.opt_A = i + 1
		} else if os.Args[i] == "-m" {
			opts.opt_M = i + 1
		} else if os.Args[i] == "-c" {
			opts.opt_C = i + 1
		} else if os.Args[i] == "-s" {
			opts.opt_S = i + 1
		}
	}
	return opts
}

type Config struct {
	Addr   string
	Role   string
	Port   string
	Passwd string
}

func main() {
	check := check_opt()
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
		go func(ip, port, passwd, role string) {
			defer wg.Done()
			addr := ip + ":" + port
			config := Make_Config("root", passwd)
			client, err := ssh.Dial("tcp", addr, config)
			if err != nil {
				log.Fatal("Failed to dial: ", err)
			}
			if opt == "-sh" {
				if check.opt_C != 0 {
					session, err := client.NewSession()
					defer session.Close()
					if err != nil {
						log.Fatal("Failed to create session: ", err)
					}
					cmd := os.Args[check.opt_C]
					if err := session.Run(cmd); err != nil {
						log.Fatal("Failed to run: " + err.Error())
					}
				}
				if (check.opt_A != 0) && (role == "agent") {
					session, err := client.NewSession()
					if err != nil {
						log.Fatal("Failed to create session: ", err)
					}
					cmd := os.Args[check.opt_A]
					if err := session.Run(cmd); err != nil {
						log.Fatal("Failed to run " + err.Error())
					}
				} else if (check.opt_M != 0) && (role == "master") {
					session, err := client.NewSession()
					if err != nil {
						log.Fatal("Failed to create session: ", err)
					}
					cmd := os.Args[check.opt_M]
					if err := session.Run(cmd); err != nil {
						log.Fatal("Failed to run " + err.Error())
					}
				} else if (check.opt_S != 0) && (role == "storage") {
					session, err := client.NewSession()
					if err != nil {
						log.Fatal("Failed to create session: ", err)
					}
					cmd := os.Args[check.opt_S]
					if err := session.Run(cmd); err != nil {
						log.Fatal("Failed to run " + err.Error())
					}
				}
			} else if opt == "-cp" {
				if (check.opt_A | check.opt_M |check.opt_C | check.opt_S) != 0 {
					fmt.Println("'-a' or '-m' can be used for '-sh' option")
					return
				}
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
		}(data[i].Addr, data[i].Port, data[i].Passwd, data[i].Role)
	}
	wg.Wait()
	fmt.Println("Complete")
}
