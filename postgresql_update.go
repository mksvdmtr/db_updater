package main

import (
	"fmt"
	"github.com/cheggaaa/pb/v3"
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

	log.Println("Drop all tables ...")

	cmd = fmt.Sprintf("PGPASSWORD=%s psql -h %s -p %s -U %s -d %s -c \"select 'drop table if exists \\\"' || tablename || '\\\" cascade;' from pg_tables where schemaname = 'public';\" | grep drop | tr -d '\"' > drop.sql || echo empty",
		c.LocalDBConfigs.Password,
		c.LocalDBConfigs.Host,
		strconv.Itoa(c.LocalDBConfigs.Port),
		c.LocalDBConfigs.Username,
		c.LocalDBConfigs.Database,
	)
	execute(cmd)
	cmd = fmt.Sprintf("PGPASSWORD=%s psql -h %s -p %s -U %s -d %s < drop.sql",
		c.LocalDBConfigs.Password,
		c.LocalDBConfigs.Host,
		strconv.Itoa(c.LocalDBConfigs.Port),
		c.LocalDBConfigs.Username,
		c.LocalDBConfigs.Database,
	)
	execute(cmd)

	os.Remove("drop.sql")

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
