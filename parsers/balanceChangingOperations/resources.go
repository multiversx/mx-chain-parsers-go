package balanceChangingOperations

type IndexedTransactionBundle struct {
	Hash               string              `json:"txHash"`
	Sender             string              `json:"sender"`
	SenderShard        uint32              `json:"senderShard"`
	Receiver           string              `json:"receiver"`
	Value              string              `json:"value"`
	Data               []byte              `json:"data,omitempty"`
	Fee                string              `json:"fee,omitempty"`
	Status             string              `json:"status"`
	TransactionReceipt *TransactionReceipt `json:"receipt,omitempty"`
}

type TransactionReceipt struct {
	Value           string `json:"value"`
	Sender          string `json:"sender"`
	Data            string `json:"data,omitempty"`
	TransactionHash string `json:"txHash"`
}

type Operation struct {
	Identifier     string                 `json:"identifier"`
	Status         OperationStatus        `json:"status"`
	Type           OperationType          `json:"type"`
	Subtype        OperationSubtype       `json:"subtype,omitempty"`
	Address        string                 `json:"address"`
	AmountValue    string                 `json:"amountValue"`
	AmountCurrency string                 `json:"amountCurrency"`
	AmountMetadata map[string]interface{} `json:"metadata,omitempty"`
}

type OperationStatus int

const (
	OperationStatusSuccess OperationStatus = iota
	OperationStatusFailure
	OperationStatusPending
)

type OperationType int
type OperationSubtype int

const (
	OperationTypeTransfer OperationType = iota
	OperationTypeFee
	OperationTypeFeeRefund
	OperationTypeFeeReward
	OperationTypeTokenManagement
)

const (
	// Subtypes for transfers
	OperationSubtypeTransferNative OperationSubtype = iota
	OperationSubtypeTransferCustomFungible
	OperationSubtypeTransferCustomSemiFungible
	OperationSubtypeTransferCustomNonFungible

	// Subtypes for fee
	OperationSubtypeFeeRegular
	OperationSubtypeFeeOfInvalidTransaction

	// Subtypes for fee refund
	OperationSubtypeFeeRefundAsReceipt
	OperationSubtypeFeeRefundAsSmartContractResult

	// Subtypes for rewards
	OperationSubtypeStakingRewards
	OperationSubtypeDelegationRewards
	OperationSubtypeDeveloperRewards

	// Subtypes for token management operations
	OperationSubtypeCustomTokenMint
	OperationSubtypeCustomTokenBurn
	OperationSubtypeCustomTokenWipe
)
