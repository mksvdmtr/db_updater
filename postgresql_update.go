package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

func (c Configs) postgresqlUpdate() {
	dumpTempPath := fmt.Sprintf("/tmp/%s-%s.bak",
		c.RemoteDBConfigs.Database,
		strconv.Itoa(time.Now().Nanosecond()),
	)
	cmd := fmt.Sprintf("PGPASSWORD=%s %s -h %s -p %s -U %s %s -Fc > %s",
		c.RemoteDBConfigs.Password,
		c.RemotePGDumpPath,
		c.RemoteDBConfigs.Host,
		strconv.Itoa(c.RemoteDBConfigs.Port),
		c.RemoteDBConfigs.Username,
		c.RemoteDBConfigs.Database,
		dumpTempPath,
	)
	fmt.Println(cmd)
	sshCmd := c.SSHClient.Cmd(cmd)
	log.Println("Creating dump ...")
	_, err := sshCmd.SmartOutput()
	handleErr(err, "Error creating postgresql dump")

	srcDump, err := configs.SFTPClient.Open(dumpTempPath)
	handleErr(err, "Cannot found source dump")
	defer srcDump.Close()
	dstDump, err := os.Create(dumpTempPath)
	handleErr(err, "Cannot create destination dump")
	defer dstDump.Close()
	_, err = io.Copy(dstDump, srcDump)
	handleErr(err, "Cannot copy source dump to destination dump")
	handleErr(c.SFTPClient.Remove(dumpTempPath), "Cannot remove %s", dumpTempPath)

	cmd = fmt.Sprintf("PGPASSWORD=%s psql -h %s -p %s -U %s -d %s -c 'DROP SCHEMA public CASCADE; "+
		"CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO postgres; GRANT ALL ON SCHEMA public TO public;'",
		c.LocalDBConfigs.Password,
		c.LocalDBConfigs.Host,
		strconv.Itoa(c.LocalDBConfigs.Port),
		c.LocalDBConfigs.Username,
		c.LocalDBConfigs.Database,
	)
	log.Println("Drop/Create schema public ...")
	execute(cmd)

	cmd = fmt.Sprintf("PGPASSWORD=%s %s -h %s -p %s -U %s -d %s -Fc -O -x %s",
		c.LocalDBConfigs.Password,
		c.LocalPGRestorePath,
		c.LocalDBConfigs.Host,
		strconv.Itoa(c.LocalDBConfigs.Port),
		c.LocalDBConfigs.Username,
		c.LocalDBConfigs.Database,
		dumpTempPath,
	)

	log.Println("Restoring db ...")
	execute(cmd)
	os.Remove(dumpTempPath)
}
