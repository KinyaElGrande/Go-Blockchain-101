package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
)

const reward = 210000

// TxInput references a previous output
type TxInput struct {
	ID        []byte
	Out       int
	Signature string
}

// TxOutput stores the coins
type TxOutput struct {
	Value  int
	PubKey string
}

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func NewCoinBaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s' ", to)
	}

	txIn := TxInput{[]byte{}, -1, data}
	txOut := TxOutput{reward, to}
	txn := Transaction{nil, []TxInput{txIn}, []TxOutput{txOut}}
	txn.SetID()

	return &txn
}

func (tx *Transaction) Iscoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	handleErr(err)

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

func (in *TxInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.Signature == unlockingData
}

func (out *TxOutput) CanbeUnlockedWith(unlockingData string) bool {
	return out.PubKey == unlockingData
}

func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTxns []Transaction
	spentTxns := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txId := hex.EncodeToString(tx.ID)

		outputs:
			for outIdx, out := range tx.Outputs {
				if spentTxns[txId] != nil {
					for _, spentOut := range spentTxns[txId] {
						if spentOut == outIdx {
							continue outputs
						}
					}
				}

				if out.CanbeUnlockedWith(address) {
					unspentTxns = append(unspentTxns, *tx)
				}
			}

			if !tx.Iscoinbase() {
				for _, in := range tx.Inputs {
					if in.CanUnlockOutputWith(address) {
						inTxId := hex.EncodeToString(in.ID)
						spentTxns[inTxId] = append(spentTxns[inTxId], in.Out)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTxns
}

func (bc *Blockchain) FindUnspentTxOutput(address string) []TxOutput {
	var unspentTxOutputs []TxOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.CanbeUnlockedWith(address) {
				unspentTxOutputs = append(unspentTxOutputs, out)
			}
		}
	}

	return unspentTxOutputs
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTransactions := bc.FindUnspentTransactions(address)
	accumulated := 0

work:
	for _, tx := range unspentTransactions {
		txnID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Outputs {
			if out.CanbeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txnID] = append(unspentOuts[txnID], outIdx)

				if accumulated >= amount {
					break work
				}
			}
		}
	}

	return accumulated, unspentOuts
}

func NewTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		errors.Errorf("Error: not enough funds")
	}

	// Build a list of inputs
	for txId, outs := range validOutputs {
		txnID, err := hex.DecodeString(txId)
		handleErr(err)

		for _, out := range outs {
			input := TxInput{txnID, out, from}
			inputs = append(inputs, input)
		}
	}

	// Build  a list of Outputs
	outputs = append(outputs, TxOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs , outputs}
	tx.SetID()

	return &tx
}
