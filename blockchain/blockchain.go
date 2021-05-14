package blockchain

import (
	"fmt"

	"github.com/dgraph-io/badger/v3"
)

const (
	genCoinbaseData = "ElGrande blockchain 101"

	dbPath = "./tmp/blocks"
)

type Blockchain struct {
	LatestBlockHash []byte
	Database        *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Db          *badger.DB
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address string) *Blockchain {

	var latestHash []byte

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	handleErr(err)

	err = db.Update(func(tx *badger.Txn) error {
		cbTxn := NewCoinBaseTX(address, genCoinbaseData)
		genesis := GenesisBlock(cbTxn)
		fmt.Println("Genesis file created")
		err = tx.Set(genesis.Hash, genesis.Serialize())
		handleErr(err)
		err = tx.Set([]byte("lh"), genesis.Hash)

		latestHash = genesis.Hash

		return err
	})
	handleErr(err)

	blockchain := Blockchain{latestHash, db}
	return &blockchain
}

func AddBlockchain(address string) *Blockchain {
	var latestHash []byte

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	handleErr(err)

	err = db.Update(func(tx *badger.Txn) error {
		item, err := tx.Get([]byte("lh"))
		handleErr(err)
		err = item.Value(func(val []byte) error {
			latestHash = val
			return nil
		})
		handleErr(err)
		return err
	})
	handleErr(err)

	bc := Blockchain{latestHash, db}
	return &bc
}

// AddBlock adds a new block on the Blockchain
func (bc *Blockchain) AddBlock(transactions []*Transaction) {
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

	newBlock := NewBlock(transactions, latestHash)

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
