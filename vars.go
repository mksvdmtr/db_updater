package main

import "github.com/helloyi/go-sshclient"

type Configs struct {
	RemoteHost       string `yaml:"remote_host"`
	RemotePort       string `yaml:"remote_port"`
	RemoteUser       string `yaml:"remote_user"`
	RemotePGDumpPath string `yaml:"remote_pg_dump_path"`
	RemoteDBConfPath string `yaml:"remote_db_conf_path"`
	RemoteEnv        string `yaml:"remote_env"`
	LocalEnv         string `yaml:"local_env"`
	LocalDBConfPath  string `yaml:"local_db_conf_path"`
	DBConfigs        DBConfigs
	SSHClient        *sshclient.Client
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
	db      Configs
)

func init() {
	configs = Configs{
		RemotePGDumpPath: "/usr/bin/pg_dump",
		RemotePort:       "22",
		RemoteEnv:        "production",
		LocalEnv:         "staging",
	}
}
