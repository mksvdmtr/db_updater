package main

import (
	"fmt"
	"github.com/cheggaaa/pb/v3"
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
	log.Println("Fetching dump ...")
	srcDumpStat, err := srcDump.Stat()
	handleErr(err, "Error getting Stat from sorce dump file")
	bar := pb.New64(srcDumpStat.Size())
	bar.Set(pb.Bytes, true)
	barReader := bar.NewProxyReader(srcDump)
	bar.Start()
	_, err = io.Copy(dstDump, barReader)
	bar.Finish()
	handleErr(err, "Cannot copy source dump to destination dump")
	handleErr(c.SFTPClient.Remove(dumpTempPath), "Cannot remove %s", dumpTempPath)

	cmd = fmt.Sprintf("mysql -h 127.0.0.1 -u %s -p%s -b %s -e 'show tables;'",
		c.LocalDBConfigs.Username,
		c.LocalDBConfigs.Password,
		c.LocalDBConfigs.Database,
	)

	cmdPrep := exec.Command("bash", "-c", cmd)
	cmdOut, err := cmdPrep.CombinedOutput()
	if len(cmdOut) > 0 {
		cmd = fmt.Sprintf("echo \"SET FOREIGN_KEY_CHECKS = 0;\" > drop.sql "+
			"&& mysqldump -h 127.0.0.1 --add-drop-table --no-data -u %s -p%s %s | grep 'DROP TABLE' >> drop.sql "+
			"&& echo \"SET FOREIGN_KEY_CHECKS = 1;\" >> drop.sql && mysql -h 127.0.0.1 -u %s -p%s %s < drop.sql "+
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

	cmd = fmt.Sprintf("gunzip < %s | mysql -h 127.0.0.1 -u %s -p%s %s",
		dumpTempPath,
		c.LocalDBConfigs.Username,
		c.LocalDBConfigs.Password,
		c.LocalDBConfigs.Database,
	)
	log.Println("Restoring dump ...")
	execute(cmd)
	os.Remove(dumpTempPath)
}
