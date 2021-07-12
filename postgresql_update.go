package main

import (
	"fmt"
	"strconv"
	"time"
)

func (c Configs) postgresqlUpdate() error {
	dumpTempName := time.Now().Nanosecond()
	cmd := fmt.Sprintf("%s -h %s -p %s -U %s %s -Fc > /tmp/%s-%s.bak",
		configs.RemotePGDumpPath,
		c.DBConfigs.Host,
		strconv.Itoa(c.DBConfigs.Port),
		c.DBConfigs.Username,
		c.DBConfigs.Database,
		c.DBConfigs.Database,
		strconv.Itoa(dumpTempName),
	)
	fmt.Println(cmd)
	return nil
}
