package main

import (
	"C"
)
import (
	"encoding/json"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-parsers-go/parsers/balanceChangingOperations"
	vmcommonParsers "github.com/multiversx/mx-chain-vm-common-go/parsers"
)

var (
	parsers = make([]balanceChangingOperations.IndexedTransferParser, 0)
	log     = logger.GetOrCreate("libraries/libparsers")
)

type indexedTransferConfig struct {
	MinGasLimit     uint64 `json:"minGasLimit"`
	GasLimitPerByte uint64 `json:"gasLimitPerByte"`
	PubkeyLength    int    `json:"pubkeyLength"`
}

func main() {
}

//export newIndexedTransferParser
func newIndexedTransferParser(configJson *C.char) int {
	configJsonString := C.GoString(configJson)

	var config indexedTransferConfig
	err := json.Unmarshal([]byte(configJsonString), &config)
	if err != nil {
		log.Error("newIndexedTransferParser(): cannot unmarshal config", err)
		return -1
	}

	pubKeyConverter, err := pubkeyConverter.NewBech32PubkeyConverter(config.PubkeyLength, log)
	if err != nil {
		log.Error("newIndexedTransferParser(): cannot create pubkey converter", err)
		return -1
	}

	parser, err := balanceChangingOperations.NewIndexedTransferParser(balanceChangingOperations.IndexedTransferParserArgs{
		PubkeyConverter: pubKeyConverter,
		CallArgsParser:  vmcommonParsers.NewCallArgsParser(),
		MinGasLimit:     config.MinGasLimit,
		GasLimitPerByte: config.GasLimitPerByte,
	})
	if err != nil {
		log.Error("newIndexedTransferParser(): cannot create parser", err)
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
		log.Error("parseIndexedTransfer(): cannot unmarshal transfer", err)
		return C.CString("")
	}

	operations, err := parsers[parserHandle].ParseTransfer(transfer)
	if err != nil {
		log.Error("parseIndexedTransfer(): cannot parse transfer", err)
		return C.CString("")
	}

	operationsJson, err := json.Marshal(operations)
	if err != nil {
		log.Error("parseIndexedTransfer(): cannot marshal operations", err)
		return C.CString("")
	}

	return C.CString(string(operationsJson))
}
