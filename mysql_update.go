package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func (c *Configs) mysqlUpdate() {
	dumpTempPath := fmt.Sprintf("/tmp/%s-%s.sql.gz",
		c.RemoteDBConfigs.Database,
		strconv.Itoa(time.Now().Nanosecond()),
	)
	cmd := fmt.Sprintf("mysqldump -u %s -p%s %s | gzip > %s",
		c.RemoteDBConfigs.Username,
		c.RemoteDBConfigs.Password,
		c.RemoteDBConfigs.Database,
		dumpTempPath,
		)

	log.Println("Creating dump ...")
	_, err := c.SSHClient.Cmd(cmd).SmartOutput()
	handleErr(err, "Error creating mysql dump")

	srcDump, err := configs.SFTPClient.Open(dumpTempPath)
	handleErr(err, "Cannot found source dump")
	defer srcDump.Close()
	dstDump, err := os.Create(dumpTempPath)
	handleErr(err, "Cannot create destination dump")
	defer dstDump.Close()
	_, err = io.Copy(dstDump, srcDump)
	handleErr(err, "Cannot copy source dump to destination dump")
	handleErr(c.SFTPClient.Remove(dumpTempPath), "Cannot remove %s", dumpTempPath)

	cmd = fmt.Sprintf("mysql -h localhost -u %s -p%s -b %s -e 'show tables;'",
		c.LocalDBConfigs.Username,
		c.LocalDBConfigs.Password,
		c.LocalDBConfigs.Database,
		)

	cmdPrep := exec.Command("bash", "-c", cmd)
	cmdOut, err := cmdPrep.CombinedOutput()
	if len(cmdOut) > 0 {
		cmd = fmt.Sprintf("echo \"SET FOREIGN_KEY_CHECKS = 0;\" > drop.sql "+
			"&& mysqldump -h localhost --add-drop-table --no-data -u %s -p%s %s | grep 'DROP TABLE' >> drop.sql "+
			"&& echo \"SET FOREIGN_KEY_CHECKS = 1;\" >> drop.sql && mysql -h localhost -u %s -p%s %s < drop.sql " +
			"&& rm drop.sql",
			c.LocalDBConfigs.Username,
			c.LocalDBConfigs.Password,
			c.LocalDBConfigs.Database,
			c.LocalDBConfigs.Username,
			c.LocalDBConfigs.Password,
			c.LocalDBConfigs.Database,
		)
		fmt.Println(cmd)
		log.Println("Drop all tables ...")
		execute(cmd)
	}

	cmd = fmt.Sprintf("gunzip < %s | mysql -h localhost -u %s -p%s %s",
		dumpTempPath,
		c.LocalDBConfigs.Username,
		c.LocalDBConfigs.Password,
		c.LocalDBConfigs.Database,
		)
	fmt.Println("Restoring dump ...")
	execute(cmd)
	os.Remove(dumpTempPath)
}
