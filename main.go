package main

import (
	"os"

	"github.com/KinyaElGrande/Go-Blockchain-101/cli"
)

func main() {
	defer os.Exit(0)
	cli := cli.CLI{}
	cli.Run()
}
