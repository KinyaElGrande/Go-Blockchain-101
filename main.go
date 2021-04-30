package main

import (
	"fmt"
	"strconv"

	"github.com/KinyaElGrande/Go-Blockchain-101/pow"
)

func main() {
	bc := pow.NewBlockchain()

	bc.AddBlock("Block 001")
	bc.AddBlock("heheh 002")

	for _, block := range bc.Blocks {
		fmt.Printf("Prev Hash : %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := pow.NewProofOfWork(block)
		fmt.Printf("POW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}

}
