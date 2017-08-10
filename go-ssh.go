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

type Opt struct {
	opt_A [2]int
	opt_M [2]int
	opt_C [2]int
	opt_S [2]int
}

type Config struct {
	Addr   string
	Role   string
	Port   string
	Passwd string
}

func endpoint(start int, arr []string) int {
	for i := start + 1; i < len(arr); i++ {
		if arr[i][0] == '-' {
			fmt.Println(arr[i], i)
			return i - 1
		}
	}
	return len(arr) - 1
}

func check_opt() Opt {
	var opts Opt
	opts.opt_A[0] = 0
	opts.opt_M[0] = 0
	opts.opt_S[0] = 0
	opts.opt_C[0] = 0
	for i := 0; i < len(os.Args); i++ {
		if os.Args[i] == "-a" {
			opts.opt_A[0] = i + 1
			opts.opt_A[1] = endpoint(i+1, os.Args)
		} else if os.Args[i] == "-m" {
			opts.opt_M[0] = i + 1
			opts.opt_M[1] = endpoint(i+1, os.Args)
		} else if os.Args[i] == "-c" {
			opts.opt_C[0] = i + 1
			opts.opt_C[1] = endpoint(i+1, os.Args)
		} else if os.Args[i] == "-s" {
			opts.opt_S[0] = i + 1
			opts.opt_S[1] = endpoint(i+1, os.Args)
		}
	}
	return opts
}

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

func cmd_run(config *ssh.ClientConfig, addr, cmd string) {
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	session, err := client.NewSession()
	defer session.Close()
	if err != nil {
		log.Fatal("Faile to create session : ", err)
	}
	if err := session.Run(cmd); err != nil {
		log.Fatal("Failed to run "+cmd, err)
	}
}
func copy_run(start_index, end_index int, role_wg *sync.WaitGroup, client *ssh.Client) {
	defer role_wg.Done()
	var cp_wg sync.WaitGroup
	for j := start_index; j < end_index; j++ {
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
				log.Fatal("Failed to scp for ", filePath, desPath)
			}
		}(os.Args[j], os.Args[end_index], client)
	}
	cp_wg.Wait()
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
			if opt == "-sh" {
				if check.opt_C[0] != 0 {
					cmd_run(config, addr, os.Args[check.opt_C[0]])
				}
				if (check.opt_A[0] != 0) && (role == "agent") {
					cmd_run(config, addr, os.Args[check.opt_A[0]])
				} else if (check.opt_M[0] != 0) && (role == "master") {
					cmd_run(config, addr, os.Args[check.opt_M[0]])
				} else if (check.opt_S[0] != 0) && (role == "storage") {
					cmd_run(config, addr, os.Args[check.opt_S[0]])
				}
			} else if opt == "-cp" {
				var role_wg sync.WaitGroup
				if check.opt_C[0] != 0 {
					role_wg.Add(1)
					client, err := ssh.Dial("tcp", addr, config)
					if err != nil {
						log.Fatal("Failed to dial: ", err)
					}
					go copy_run(check.opt_C[0], check.opt_C[1], &role_wg, client)
				}
				if (check.opt_A[0] != 0) && (role == "agent") {
					role_wg.Add(1)
					client, err := ssh.Dial("tcp", addr, config)
					if err != nil {
						log.Fatal("Failed to dial: ", err)
					}
					go copy_run(check.opt_A[0], check.opt_A[1], &role_wg, client)
				} else if (check.opt_M[0] != 0) && (role == "master") {
					role_wg.Add(1)
					client, err := ssh.Dial("tcp", addr, config)
					if err != nil {
						log.Fatal("Failed to dial: ", err)
					}
					go copy_run(check.opt_M[0], check.opt_M[1], &role_wg, client)
				} else if (check.opt_S[0] != 0) && (role == "storage") {
					role_wg.Add(1)
					client, err := ssh.Dial("tcp", addr, config)
					if err != nil {
						log.Fatal("Failed to dial: ", err)
					}
					go copy_run(check.opt_S[0], check.opt_S[1], &role_wg, client)
				}
				role_wg.Wait()
			}
		}(data[i].Addr, data[i].Port, data[i].Passwd, data[i].Role)
	}
	wg.Wait()
	fmt.Println("Complete")
}
