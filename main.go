package main

import (
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
	configsFileData, err := ioutil.ReadFile("config.yml")
	handleErr(err, "Cannot found config file")
	err = yaml.Unmarshal(configsFileData, &configs)
	handleErr(err, "Cannot unmarshal config file")
	addr := fmt.Sprintf("%s:%s", configs.RemoteHost, configs.RemotePort)
	configs.SSHClient, err = sshclient.DialWithKey(addr, configs.RemoteUser, "/home/mda/.ssh/id_rsa2")
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
	fmt.Println(configs.RemoteDBConfigs)
	localDBConfigsFileData, err := ioutil.ReadFile(configs.LocalDBConfPath)
	handleErr(err, "Cannot open local_db_conf_path")
	localDBConfigs := make(map[string]DBConfigs, 0)
	handleErr(yaml.Unmarshal([]byte(localDBConfigsFileData), &localDBConfigs), "Cannot unmarshal yaml")
	localDBParsedConfig := localDBConfigs[configs.LocalEnv]
	handleErr(mapstructure.Decode(localDBParsedConfig, &configs.LocalDBConfigs), "Cannot unmarshal %s", localDBConfigs)
	fmt.Println(configs.LocalDBConfigs)
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
