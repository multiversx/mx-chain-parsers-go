package balanceChangingOperations

type Operation struct {
	Status         OperationStatus        `json:"status"`
	Type           OperationType          `json:"type"`
	Subtype        OperationSubtype       `json:"subtype,omitempty"`
	Address        string                 `json:"address"`
	AmountValue    string                 `json:"amountValue"`
	AmountType     AmountType             `json:"amountType"`
	AmountCurrency string                 `json:"amountCurrency,omitempty"`
	AmountMetadata map[string]interface{} `json:"metadata,omitempty"`
	Direction      OperationDirection     `json:"direction"`
}

type OperationStatus string

const (
	OperationStatusSuccess OperationStatus = "success"
	OperationStatusFailure                 = "failure"
	OperationStatusPending                 = "pending"
)

func (operationStatus OperationStatus) String() string {
	return string(operationStatus)
}

type OperationType string
type OperationSubtype string

const (
	OperationTypeTransfer        OperationType = "transfer"
	OperationTypeFee                           = "fee"
	OperationTypeFeeRefund                     = "feeRefund"
	OperationTypeReward                        = "reward"
	OperationTypeTokenManagement               = "tokenManagement"
)

func (operationType OperationType) String() string {
	return string(operationType)
}

const (
	// Subtypes for transfers
	OperationSubtypeTransferNative             OperationSubtype = "transferNative"
	OperationSubtypeTransferCustomFungible                      = "transferCustomFungible"
	OperationSubtypeTransferCustomSemiFungible                  = "transferCustomSemiFungible"
	OperationSubtypeTransferCustomNonFungible                   = "transferCustomNonFungible"

	// Subtypes for fee
	OperationSubtypeFeeRegular              = "feeRegular"
	OperationSubtypeFeeOfInvalidTransaction = "feeOfInvalidTransaction"

	// Subtypes for fee refund
	OperationSubtypeFeeRefundAsReceipt             = "feeRefundAsReceipt"
	OperationSubtypeFeeRefundAsSmartContractResult = "feeRefundAsSmartContractResult"

	// Subtypes for rewards
	OperationSubtypeStakingRewards    = "stakingRewards"
	OperationSubtypeDelegationRewards = "delegationRewards"
	OperationSubtypeDeveloperRewards  = "developerRewards"

	// Subtypes for token management operations
	OperationSubtypeCustomTokenMint = "customTokenMint"
	OperationSubtypeCustomTokenBurn = "customTokenBurn"
	OperationSubtypeCustomTokenWipe = "customTokenWipe"
)

func (operationSubtype OperationSubtype) String() string {
	return string(operationSubtype)
}

type AmountType string

const (
	AmountTypeNative             AmountType = "native"
	AmountTypeCustomFungible                = "customFungible"
	AmountTypeCustomSemiFungible            = "customSemiFungible"
	AmountTypeCustomNonFungible             = "customNonFungible"
)

func (amountType AmountType) String() string {
	return string(amountType)
}

type OperationDirection string

const (
	OperationDirectionCredit OperationDirection = "credit"
	OperationDirectionDebit  OperationDirection = "debit"
)

func (direction OperationDirection) String() string {
	return string(direction)
}
