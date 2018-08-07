package main

import (
	"os"
	"os/exec"
)

// Execute commands via bash.
func BashExec(command string) {
	cmd := exec.Command("bash", "-c", command)
	err := cmd.Run()
	if err != nil {
		L.Fatalf("Failed running command: %s - %s\n", command, err)
	}

}

// Get size of a file.
func FileSize(file string) int {
	info, err := os.Stat(file)
	if err != nil {
		L.Fatalf("Could not stat file %s: %s\n", file, err)
	}
	return int(info.Size())
}
