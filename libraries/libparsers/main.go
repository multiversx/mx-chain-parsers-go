package main

import (
	"C"
)
import (
	"encoding/json"

	"log"

	"github.com/multiversx/mx-chain-parsers-go/parsers/balanceChangingOperations"
)

var (
	parsers = make([]balanceChangingOperations.IndexedTransactionParser, 0)
)

func main() {
}

//export newIndexedTransactionParser
func newIndexedTransactionParser(configJson *C.char) int {
	configJsonString := C.GoString(configJson)

	var config balanceChangingOperations.IndexedTransactionParserConfig
	err := json.Unmarshal([]byte(configJsonString), &config)
	if err != nil {
		log.Println("newIndexedTransactionParser(): cannot unmarshal config", err)
		return -1
	}

	parser, err := balanceChangingOperations.NewIndexedTransactionParser(config)
	if err != nil {
		log.Println("newIndexedTransactionParser(): cannot create parser", err)
		return -1
	}

	parsers = append(parsers, *parser)
	return len(parsers) - 1
}

//export parseIndexedTransaction
func parseIndexedTransaction(parserHandle int, transactionJson *C.char) *C.char {
	transactionJsonString := C.GoString(transactionJson)

	var transaction balanceChangingOperations.IndexedTransaction
	err := json.Unmarshal([]byte(transactionJsonString), &transaction)
	if err != nil {
		log.Println("parseTransaction(): cannot unmarshal transaction", err)
		return C.CString("")
	}

	operations, err := parsers[parserHandle].ParseTransaction(transaction)
	if err != nil {
		log.Println("parseTransaction(): cannot parse transaction", err)
		return C.CString("")
	}

	operationsJson, err := json.Marshal(operations)
	if err != nil {
		log.Println("parseTransaction(): cannot marshal operations", err)
		return C.CString("")
	}

	return C.CString(string(operationsJson))
}
