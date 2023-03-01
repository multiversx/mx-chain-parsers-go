package balanceChangingOperations

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-parsers-go/parsers"
)

type indexedTransactionBundleParser struct {
	config IndexedTransactionBundleParserConfig
}

// NewIndexedTransactionBundleParser creates a new indexedTransactionBundleParser
func NewIndexedTransactionBundleParser(config IndexedTransactionBundleParserConfig) (*indexedTransactionBundleParser, error) {
	return &indexedTransactionBundleParser{
		config: config,
	}, nil
}

// ParseBundle parses a transaction into a list of balance-changing operations
func (parser *indexedTransactionBundleParser) ParseBundle(bundle IndexedTransactionBundle) ([]Operation, error) {
	if parser.isStakingRewards(&bundle) {
		return parser.parseStakingRewardsTransaction(&bundle)
	}
	if parser.isInvalidTransaction(&bundle) {
		return parser.parseInvalidTransaction(&bundle)
	}
	if parser.isSmartContractResult(&bundle) {
		return parser.parseSmartContractResult(&bundle)
	}

	return parser.parseRegularTransaction(&bundle)
}

func (parser *indexedTransactionBundleParser) isStakingRewards(bundle *IndexedTransactionBundle) bool {
	isFromMetachain := bundle.SenderShard == core.MetachainShardId
	isNonZero := parsers.IsNonZeroAmount(bundle.Value)
	hasNoData := len(bundle.Data) == 0

	return isFromMetachain && isNonZero && hasNoData
}

func (parser *indexedTransactionBundleParser) parseStakingRewardsTransaction(bundle *IndexedTransactionBundle) ([]Operation, error) {
	return []Operation{
		{
			Type:        OperationTypeReward,
			Subtype:     OperationSubtypeStakingRewards,
			Status:      OperationStatusSuccess,
			Address:     bundle.Receiver,
			AmountValue: bundle.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionCredit,
		},
	}, nil
}

func (parser *indexedTransactionBundleParser) isInvalidTransaction(bundle *IndexedTransactionBundle) bool {
	return bundle.Status == TransactionStatusInvalid
}

func (parser *indexedTransactionBundleParser) parseInvalidTransaction(bundle *IndexedTransactionBundle) ([]Operation, error) {
	operations := make([]Operation, 0)

	operations = append(operations, Operation{
		Type:        OperationTypeFee,
		Subtype:     OperationSubtypeFeeOfInvalidTransaction,
		Status:      OperationStatusSuccess,
		Address:     bundle.Sender,
		AmountValue: bundle.Fee,
		AmountType:  AmountTypeNative,
		Direction:   OperationDirectionDebit,
	})

	if parsers.IsNonZeroAmount(bundle.Value) {
		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusFailure,
			Address:     bundle.Sender,
			AmountValue: bundle.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionDebit,
		})

		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusFailure,
			Address:     bundle.Receiver,
			AmountValue: bundle.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionCredit,
		})
	}

	return operations, nil
}

func (parser *indexedTransactionBundleParser) isSmartContractResult(bundle *IndexedTransactionBundle) bool {
	return bundle.Type == TransactionTypeSmartContractResult
}

func (parser *indexedTransactionBundleParser) parseSmartContractResult(bundle *IndexedTransactionBundle) ([]Operation, error) {
	operations := make([]Operation, 0)

	if parsers.IsNonZeroAmount(bundle.Value) {
		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusSuccess,
			Address:     bundle.Sender,
			AmountValue: bundle.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionDebit,
		})

		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusSuccess,
			Address:     bundle.Receiver,
			AmountValue: bundle.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionCredit,
		})
	}

	return operations, nil
}

func (parser *indexedTransactionBundleParser) parseRegularTransaction(bundle *IndexedTransactionBundle) ([]Operation, error) {
	operations := make([]Operation, 0)

	operations = append(operations, Operation{
		Type:        OperationTypeFee,
		Subtype:     OperationSubtypeFeeRegular,
		Status:      OperationStatusSuccess,
		Address:     bundle.Sender,
		AmountValue: bundle.Fee,
		AmountType:  AmountTypeNative,
		Direction:   OperationDirectionDebit,
	})

	if parsers.IsNonZeroAmount(bundle.Value) {
		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusSuccess,
			Address:     bundle.Sender,
			AmountValue: bundle.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionDebit,
		})

		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusSuccess,
			Address:     bundle.Receiver,
			AmountValue: bundle.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionCredit,
		})
	}

	return operations, nil
}
