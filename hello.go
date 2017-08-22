package main

import (
    "golang.org/x/crypto/ssh"
    "github.com/tmc/scp"
    "os"
    "log"
    "fmt"
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



func main() {
    args := os.Args[1:]

    master_ip := args[0]
    master_port := args[1]
    master_password := args[2]
    files_path_list := args[3:len(args)-1]
    des_path := args[len(args)-1]


    config := Make_Config("root",master_password)
    client, e := ssh.Dial("tcp",master_ip+":"+master_port,config)
    if e != nil {
    	log.Fatal("Failed to dial: ", e)
    }

    var wg sync.WaitGroup
    
    for _ , file := range files_path_list {
        wg.Add(1)
        go func() {
	    session, err := client.NewSession()
	    defer session.Close()
	    if err != nil {
	        log.Fatal("Failed to create Session")
	    }

	    defer wg.Done()    	
	    err = scp.CopyPath(file,des_path,session)
	    fmt.Println(err)
	    if(err != nil) {
	        log.Fatal("Fail copy")
	    }    	
	}()
    }



    /*
    for _ , file := range file_path {
    	session, err := client.NewSession()
	defer session.Close()
	if err != nil {
		log.Fatal("Failed to create Session")
	}
    		
    	//defer wg.Done()    	
    	err = scp.CopyPath(file,des_path,session)
    	fmt.Println(err)
    	if(err != nil) {
    		log.Fatal("Fail copy")
    	}    	
    }*/
    
    
    wg.Wait()
    fmt.Println("Complete")
}