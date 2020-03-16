package main

import (
	"os/exec"
)

// Check for the presence of commands we depend on.
func CheckDependencies() {

	checkBash()
	checkMinimap2()
	checkRacon()
}

// Check for bash.
func checkBash() {
	cmd := exec.Command("bash", "--help")
	err := cmd.Run()
	if err != nil {
		L.Fatalf("The bash command was not found in the path: %s\n", err)
	}
}

// Checks for minimap2.
func checkMinimap2() {
	BashExec("minimap2 -h")
}

// Check for racon.
func checkRacon() {
	BashExec("racon -h reads overalps raw cons")
}
