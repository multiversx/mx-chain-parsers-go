package balanceChangingOperations

// PubkeyConverter can convert public key bytes to/from a human readable form
type PubkeyConverter interface {
	Decode(humanReadable string) ([]byte, error)
	IsInterfaceNil() bool
}
