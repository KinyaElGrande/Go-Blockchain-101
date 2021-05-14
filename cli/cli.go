package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/KinyaElGrande/Go-Blockchain-101/blockchain"
)

type CLI struct{}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("getbalance -address ADDRESS - get balance for address")
	fmt.Println("createBlockchain -address ADDRESS creates a blockchain and rewards the miner")
	fmt.Println("listblocks -prints all the blocks in the blockchain")
	fmt.Println("send -from FROM -to TO -amount AMOUNT of tokens from one address to another")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) createBlockchain(address string) {
	newBlockChain := blockchain.CreateBlockchain(address)
	newBlockChain.Database.Close()
	fmt.Println("Created a new BlockChain successfully")
}

func (cli *CLI) printBlocks() {
	bc := blockchain.CreateBlockchain("")
	defer bc.Database.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("Prev Hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash : %x\n", block.Hash)
		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("POW : %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	printBlocksCmd := flag.NewFlagSet("listblocks", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance from")
	createBLockchainAddress := createBlockchainCmd.String("address", "", "the address to send reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination  wallet Address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listblocks":
		err := printBlocksCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBLockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockchain(*createBLockchainAddress)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if printBlocksCmd.Parsed() {
		cli.printBlocks()
	}
}

func (cli *CLI) getBalance(address string) {
	bc := blockchain.AddBlockchain(address)
	defer bc.Database.Close()

	balance := 0
	unspentTxOutputs := bc.FindUnspentTxOutput(address)

	for _, out := range unspentTxOutputs {
		balance += out.Value
	}

	fmt.Printf("balance of '%s: %d'\n", address, balance)
}

func (cli *CLI) send(from, to string, amount int) {
	bc := blockchain.AddBlockchain(from)
	defer bc.Database.Close()

	txn := blockchain.NewTransaction(from, to, amount, bc)

	bc.AddBlock([]*blockchain.Transaction{txn})
	fmt.Println("Token has been transferred Successfully!")
}
