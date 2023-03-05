package balanceChangingOperations

type IndexedTransfer struct {
	Hash               string              `json:"txHash"`
	Timestamp          uint64              `json:"timestamp"`
	Round              uint64              `json:"round"`
	Type               string              `json:"type"`
	Sender             string              `json:"sender"`
	SenderShard        uint32              `json:"senderShard"`
	Receiver           string              `json:"receiver"`
	Value              string              `json:"value"`
	Data               []byte              `json:"data,omitempty"`
	GasPrice           uint64              `json:"gasPrice"`
	Fee                string              `json:"fee,omitempty"`
	Status             string              `json:"status"`
	TransactionReceipt *TransactionReceipt `json:"receipt,omitempty"`
}

func (indexedTransfer IndexedTransfer) isSmartContractResult() bool {
	return indexedTransfer.Type == TransferTypeSmartContractResult
}

func (indexedTransfer IndexedTransfer) isInvalid() bool {
	return indexedTransfer.Status == TransferStatusInvalid
}

type TransactionReceipt struct {
	Value           string `json:"value"`
	Sender          string `json:"sender"`
	Data            string `json:"data,omitempty"`
	TransactionHash string `json:"txHash"`
}
