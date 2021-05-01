package main

import (
	"github.com/KinyaElGrande/Go-Blockchain-101/blockchain"
	"github.com/KinyaElGrande/Go-Blockchain-101/cli"
)

func main() {
	bc := blockchain.NewBlockchain()
	defer bc.Database.Close()
	cli := cli.CLI{bc}
	cli.Run()
}
