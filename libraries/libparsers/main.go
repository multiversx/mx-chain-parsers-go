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
	parsers = make([]balanceChangingOperations.IndexedTransferParser, 0)
)

func main() {
}

//export newIndexedTransferParser
func newIndexedTransferParser(configJson *C.char) int {
	configJsonString := C.GoString(configJson)

	var config balanceChangingOperations.IndexedTransferParserArgs
	err := json.Unmarshal([]byte(configJsonString), &config)
	if err != nil {
		log.Println("newIndexedTransferParser(): cannot unmarshal config", err)
		return -1
	}

	parser, err := balanceChangingOperations.NewIndexedTransferParser(config)
	if err != nil {
		log.Println("newIndexedTransferParser(): cannot create parser", err)
		return -1
	}

	parsers = append(parsers, *parser)
	return len(parsers) - 1
}

//export parseIndexedTransfer
func parseIndexedTransfer(parserHandle int, transferJson *C.char) *C.char {
	transferJsonString := C.GoString(transferJson)

	var transfer balanceChangingOperations.IndexedTransfer
	err := json.Unmarshal([]byte(transferJsonString), &transfer)
	if err != nil {
		log.Println("parseIndexedTransfer(): cannot unmarshal transfer", err)
		return C.CString("")
	}

	operations, err := parsers[parserHandle].ParseTransfer(transfer)
	if err != nil {
		log.Println("parseIndexedTransfer(): cannot parse transfer", err)
		return C.CString("")
	}

	operationsJson, err := json.Marshal(operations)
	if err != nil {
		log.Println("parseIndexedTransfer(): cannot marshal operations", err)
		return C.CString("")
	}

	return C.CString(string(operationsJson))
}
