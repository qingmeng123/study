package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	cmd := exec.Command("rabbitmqctl", "list_connections")

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error executing rabbitmqctl: %v", err)
	}

	fmt.Println(string(output))
}
