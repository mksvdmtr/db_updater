package main

import (
	"github.com/helloyi/go-sshclient"
	"github.com/pkg/sftp"
)

type Configs struct {
	RemoteHost         string `yaml:"remote_host"`
	RemotePort         string `yaml:"remote_port"`
	RemoteUser         string `yaml:"remote_user"`
	RemotePGDumpPath   string `yaml:"remote_pg_dump_path"`
	RemoteDBConfPath   string `yaml:"remote_db_conf_path"`
	RemoteEnv          string `yaml:"remote_env"`
	LocalEnv           string `yaml:"local_env"`
	LocalDBConfPath    string `yaml:"local_db_conf_path"`
	LocalPGRestorePath string `yaml:"local_pg_restore_path"`
	RemoteDBConfigs    DBConfigs
	LocalDBConfigs     DBConfigs
	SSHClient          *sshclient.Client
	SFTPClient         *sftp.Client
}

type DBConfigs struct {
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Adapter  string `yaml:"adapter"`
}

var (
	configs Configs
	configUsage = `
-c /path/to/config.yml
	
available parameters:
	
	remote_host: "host"
	remote_port: "22"
	remote_user: "user"
	remote_pg_dump_path: "/usr/bin/pg_dump"
	remote_db_conf_path: "/data/user/app/shared/config/database.yml"
	remote_env: "production"
	local_env: "staging"
	local_db_conf_path: "/home/user/app/shared/config/database.yml"
	local_pg_restore_path: "/usr/bin/pg_restore"

`
)

func init() {
	configs = Configs{
		RemotePGDumpPath: "/usr/bin/pg_dump",
		RemotePort:       "22",
		RemoteEnv:        "production",
		LocalEnv:         "staging",
		LocalPGRestorePath: "/usr/bin/pg_restore",
	}

}
