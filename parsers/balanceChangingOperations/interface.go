package balanceChangingOperations

// PubkeyConverter can convert public key bytes to/from a human readable form
type PubkeyConverter interface {
	Decode(humanReadable string) ([]byte, error)
	IsInterfaceNil() bool
}

// CallArgsParser can parse the arguments of a smart contract call
type CallArgsParser interface {
	ParseData(data string) (string, [][]byte, error)
	ParseArguments(data string) ([][]byte, error)
	IsInterfaceNil() bool
}
