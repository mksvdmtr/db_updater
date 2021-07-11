package main

import (
	"fmt"
	sshclient "github.com/helloyi/go-sshclient"
	"github.com/pkg/sftp"
	"io"
	"log"
	"os"
)

func main()  {

	client, err := sshclient.DialWithKey("rs0.ru:22", "root", "/home/mda/.ssh/id_rsa2")
	if err != nil {
		log.Println(err)
	}

	defer client.Close()
	if err != nil {
		fmt.Println(err)
	}

	out, err := client.Cmd("ls -l").Output()
	if err != nil {
		log.Println(err)
	}

	fmt.Println(string(out))


	out, err = client.Cmd("cat /data/stage1/app/shared/config/database.yml").Output()
	if err != nil {
		log.Println(err)
	}

	fmt.Println(string(out))
	cl := client.UnderlyingClient()
	sftpCl, err := sftp.NewClient(cl)


	srcFile, err := sftpCl.Open("/data/tothemoon/tothemoon.sql")
	if err != nil {
		log.Println(err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create("/tmp/dump.sql")
	if err != nil {
		return
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)

}