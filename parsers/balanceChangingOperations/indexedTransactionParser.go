package balanceChangingOperations

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-parsers-go/parsers"
)

// todo sending value to non payable 420cf361f54ffa9370f380aae6205b2d9e2a0c575da6dc7a25f9185d0a2d9268 (devnet)
// data + logs?

type IndexedTransactionParser struct {
	config IndexedTransactionParserConfig
}

// NewIndexedTransactionParser creates a new indexedTransactionParser
func NewIndexedTransactionParser(config IndexedTransactionParserConfig) (*IndexedTransactionParser, error) {
	return &IndexedTransactionParser{
		config: config,
	}, nil
}

// ParseTransaction parses a transaction into a list of balance-changing operations
func (parser *IndexedTransactionParser) ParseTransaction(transaction IndexedTransaction) ([]Operation, error) {
	if parser.isStakingRewards(&transaction) {
		return parser.parseStakingRewardsTransaction(&transaction)
	}
	if parser.isInvalidTransaction(&transaction) {
		return parser.parseInvalidTransaction(&transaction)
	}
	if parser.isSmartContractResult(&transaction) {
		return parser.parseSmartContractResult(&transaction)
	}

	return parser.parseRegularTransaction(&transaction)
}

func (parser *IndexedTransactionParser) isStakingRewards(transaction *IndexedTransaction) bool {
	isFromMetachain := transaction.SenderShard == core.MetachainShardId
	isNonZero := parsers.IsNonZeroAmount(transaction.Value)
	hasNoData := len(transaction.Data) == 0

	return isFromMetachain && isNonZero && hasNoData
}

func (parser *IndexedTransactionParser) parseStakingRewardsTransaction(transaction *IndexedTransaction) ([]Operation, error) {
	return []Operation{
		{
			Type:        OperationTypeReward,
			Subtype:     OperationSubtypeStakingRewards,
			Status:      OperationStatusSuccess,
			Address:     transaction.Receiver,
			AmountValue: transaction.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionCredit,
		},
	}, nil
}

func (parser *IndexedTransactionParser) isInvalidTransaction(transaction *IndexedTransaction) bool {
	return transaction.Status == TransactionStatusInvalid
}

func (parser *IndexedTransactionParser) parseInvalidTransaction(transaction *IndexedTransaction) ([]Operation, error) {
	operations := make([]Operation, 0)

	operations = append(operations, Operation{
		Type:        OperationTypeFee,
		Subtype:     OperationSubtypeFeeOfInvalidTransaction,
		Status:      OperationStatusSuccess,
		Address:     transaction.Sender,
		AmountValue: transaction.Fee,
		AmountType:  AmountTypeNative,
		Direction:   OperationDirectionDebit,
	})

	if parsers.IsNonZeroAmount(transaction.Value) {
		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusFailure,
			Address:     transaction.Sender,
			AmountValue: transaction.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionDebit,
		})

		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusFailure,
			Address:     transaction.Receiver,
			AmountValue: transaction.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionCredit,
		})
	}

	return operations, nil
}

func (parser *IndexedTransactionParser) isSmartContractResult(transaction *IndexedTransaction) bool {
	return transaction.Type == TransactionTypeSmartContractResult
}

func (parser *IndexedTransactionParser) parseSmartContractResult(transaction *IndexedTransaction) ([]Operation, error) {
	operations := make([]Operation, 0)

	if parsers.IsNonZeroAmount(transaction.Value) {
		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusSuccess,
			Address:     transaction.Sender,
			AmountValue: transaction.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionDebit,
		})

		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusSuccess,
			Address:     transaction.Receiver,
			AmountValue: transaction.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionCredit,
		})
	}

	return operations, nil
}

func (parser *IndexedTransactionParser) parseRegularTransaction(transaction *IndexedTransaction) ([]Operation, error) {
	operations := make([]Operation, 0)

	operations = append(operations, Operation{
		Type:        OperationTypeFee,
		Subtype:     OperationSubtypeFeeRegular,
		Status:      OperationStatusSuccess,
		Address:     transaction.Sender,
		AmountValue: transaction.Fee,
		AmountType:  AmountTypeNative,
		Direction:   OperationDirectionDebit,
	})

	if parsers.IsNonZeroAmount(transaction.Value) {
		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusSuccess,
			Address:     transaction.Sender,
			AmountValue: transaction.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionDebit,
		})

		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusSuccess,
			Address:     transaction.Receiver,
			AmountValue: transaction.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionCredit,
		})
	}

	return operations, nil
}
