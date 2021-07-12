package main

import (
	"os"
	"os/exec"
)

func execute(cmd string) {
	cmdPrep := exec.Command("bash", "-c", cmd)
	cmdPrep.Stdout = os.Stdout
	cmdPrep.Stderr = os.Stderr
	err := cmdPrep.Start()
	handleErr(err, "Error executing %s", cmd)
	err = cmdPrep.Wait()
	handleErr(err, "Error executing %s", cmd)
}
