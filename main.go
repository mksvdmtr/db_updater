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
	client, err := sshclient.DialWithKey(addr, configs.RemoteUser, "/home/mda/.ssh/id_rsa2")
	handleErr(err, "Cannot ssh dial")
	defer client.Close()
	cl := client.UnderlyingClient()
	sftpCl, err := sftp.NewClient(cl)
	handleErr(err, "Cannot create sftp client")
	srcFile, err := sftpCl.Open(configs.RemoteDBConfPath)
	handleErr(err, "Cannot found remote_db_conf_path")
	defer srcFile.Close()
	dstFile, err := ioutil.TempFile("/tmp", "database.yml-")
	handleErr(err, "Canot create %s temp file", dstFile.Name())
	defer dstFile.Close()
	defer os.Remove(dstFile.Name())
	dbconfigs := make(map[string]DBConfigs, 0)
	_, err = io.Copy(dstFile, srcFile)
	handleErr(err, "Cannot copy %s to %s file", srcFile.Name(), dstFile.Name())
	dbConfigsFileData, err := ioutil.ReadFile(dstFile.Name())
	handleErr(err, "Cannot read %s", dstFile.Name())
	err = yaml.Unmarshal([]byte(dbConfigsFileData), &dbconfigs)
	handleErr(err, "Cannot unmarshal %s", dbconfigs)
	dbParsedConfig := dbconfigs[configs.RemoteEnv]
	err = mapstructure.Decode(dbParsedConfig, &db.DBConfigs)
	handleErr(err, "Cannot convert map to struct")
	fmt.Println(db.DBConfigs)

	switch dbParsedConfig.Adapter {
	case "postgresql":
		handleErr(db.postgresqlUpdate(), "Failed create dump")
	case "mysql2":
		//db.postgresqlUpdate()
	default:
		handleErr(errors.New("adapter error"), "Database is unrecognized")
	}
}
