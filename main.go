package main

import (
	"flag"
	"fmt"
	"github.com/helloyi/go-sshclient"
	"github.com/mitchellh/mapstructure"
	_ "github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	yaml "gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
)

func main() {
	homeDir, err := os.UserHomeDir()
	configFile := flag.String("c", "config.yml", "-c /path/to/config.yml")
	sshKeyString := fmt.Sprintf("%s/.ssh/id_rsa", homeDir)
	privateSSHKey := flag.String("ssh-key", sshKeyString, "-ssh-key /path/to/.ssh/id_rsa")
	flag.Parse()
	handleErr(err, "Error getting user home dir")
	configsFileData, err := ioutil.ReadFile(*configFile)
	handleErr(err, "Cannot found config file")
	err = yaml.Unmarshal(configsFileData, &configs)
	handleErr(err, "Cannot unmarshal config file")
	addr := fmt.Sprintf("%s:%s", configs.RemoteHost, configs.RemotePort)
	configs.SSHClient, err = sshclient.DialWithKey(addr, configs.RemoteUser, *privateSSHKey)
	handleErr(err, "Cannot ssh dial")
	defer configs.SSHClient.Close()
	cl := configs.SSHClient.UnderlyingClient()
	configs.SFTPClient, err = sftp.NewClient(cl)
	handleErr(err, "Cannot create sftp client")
	srcFile, err := configs.SFTPClient.Open(configs.RemoteDBConfPath)
	handleErr(err, "Cannot found remote_db_conf_path")
	defer srcFile.Close()
	dstFile, err := ioutil.TempFile("/tmp", "database.yml-")
	handleErr(err, "Canot create %s temp file", dstFile.Name())
	defer dstFile.Close()
	defer os.Remove(dstFile.Name())
	remoteDBConfigs := make(map[string]DBConfigs, 0)
	_, err = io.Copy(dstFile, srcFile)
	handleErr(err, "Cannot copy %s to %s file", srcFile.Name(), dstFile.Name())
	remoteDBConfigsFileData, err := ioutil.ReadFile(dstFile.Name())
	handleErr(err, "Cannot read %s", dstFile.Name())
	handleErr(yaml.Unmarshal([]byte(remoteDBConfigsFileData), &remoteDBConfigs), "Cannot unmarshal %s", remoteDBConfigs)
	remoteDBParsedConfig := remoteDBConfigs[configs.RemoteEnv]
	err = mapstructure.Decode(remoteDBParsedConfig, &configs.RemoteDBConfigs)
	handleErr(err, "Cannot convert map to struct")
	localDBConfigsFileData, err := ioutil.ReadFile(configs.LocalDBConfPath)
	handleErr(err, "Cannot open local_db_conf_path")
	localDBConfigs := make(map[string]DBConfigs, 0)
	handleErr(yaml.Unmarshal([]byte(localDBConfigsFileData), &localDBConfigs), "Cannot unmarshal yaml")
	localDBParsedConfig := localDBConfigs[configs.LocalEnv]
	handleErr(mapstructure.Decode(localDBParsedConfig, &configs.LocalDBConfigs), "Cannot unmarshal %s", localDBConfigs)
	if remoteDBParsedConfig.Adapter != localDBParsedConfig.Adapter {
		handleErr(errors.New("Adapters mismatch"), "Remote adapter and local adapter not match")
	}
	switch remoteDBParsedConfig.Adapter {
	case "postgresql":
		configs.postgresqlUpdate()
	case "mysql2":
		configs.mysqlUpdate()
	default:
		handleErr(errors.New("adapter error"), "Database is unrecognized")
	}

	configs.SFTPClient.Close()
	configs.SSHClient.Close()
}
