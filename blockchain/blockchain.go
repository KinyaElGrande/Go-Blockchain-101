package blockchain

import (
	"fmt"
	"os"

	"github.com/dgraph-io/badger/v3"
)

type Blockchain struct {
	LatestBlockHash []byte
	Database        *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Db          *badger.DB
}

// NewBlockchain opens a connection to the database
func NewBlockchain() *Blockchain {
	var latestHash []byte

	cwd, err := os.Getwd()

	if err != nil {
		fmt.Println(err)
	}

	opts := badger.DefaultOptions(cwd)

	db, err := badger.Open(opts)
	handleErr(err)

	// Check if there is already a blockchain
	err = db.Update(func(tx *badger.Txn) error {
		// lh key is a key to LatestBlockHash
		if _, err := tx.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain Found")
			genesis := GenesisBlock()
			err = tx.Set(genesis.Hash, genesis.Serialize())
			handleErr(err)
			err = tx.Set([]byte("lh"), genesis.Hash)

			latestHash = genesis.Hash

			return err
		} else {
			item, err := tx.Get([]byte("lh"))
			handleErr(err)
			err = item.Value(func(val []byte) error {
				latestHash = val
				return nil
			})
			handleErr(err)
			return err
		}

	})
	handleErr(err)

	blockchain := Blockchain{latestHash, db}
	return &blockchain
}

// AddBlock adds a new block on the Blockchain
func (bc *Blockchain) AddBlock(data string) {
	var latestHash []byte

	err := bc.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		handleErr(err)
		err = item.Value(func(val []byte) error {
			latestHash = val
			return nil
		})
		handleErr(err)
		return err
	})
	handleErr(err)

	newBlock := NewBlock(data, latestHash)

	err = bc.Database.Update(func(tx *badger.Txn) error {
		err := tx.Set(newBlock.Hash, newBlock.Serialize())
		handleErr(err)
		err = tx.Set([]byte("lh"), newBlock.Hash)

		bc.LatestBlockHash = newBlock.Hash
		return err
	})
	handleErr(err)
}

func (bc *Blockchain) Iterator() *BlockChainIterator {
	bcIterator := BlockChainIterator{bc.LatestBlockHash, bc.Database}

	return &bcIterator
}

// Next returns the next block from the BlockChain
func (bi *BlockChainIterator) Next() *Block {
	var block *Block
	
	err := bi.Db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(bi.CurrentHash)
		handleErr(err)

		err = item.Value(func(val []byte) error {
			block = Deserialize(val)
		return nil
		})
		handleErr(err)
		return err
	})
	handleErr(err)

	bi.CurrentHash = block.PrevBlockHash

	return block
}