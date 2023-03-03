package balanceChangingOperations

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-parsers-go/parsers"
)

type IndexedTransactionParserArgs struct {
	PubkeyConverter PubkeyConverter
	MinGasLimit     uint64
}

type IndexedTransactionParser struct {
	pubkeyConverter PubkeyConverter
	minGasLimit     uint64
}

// NewIndexedTransactionParser creates a new IndexedTransactionParser
func NewIndexedTransactionParser(args IndexedTransactionParserArgs) (*IndexedTransactionParser, error) {
	if check.IfNil(args.PubkeyConverter) {
		return nil, errNilPubkeyConverter
	}
	if args.MinGasLimit == 0 {
		return nil, errBadMinGasLimit
	}

	return &IndexedTransactionParser{
		pubkeyConverter: args.PubkeyConverter,
		minGasLimit:     args.MinGasLimit,
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

	isSendingValueToNonPayableContract, err := parser.isSendingValueToNonPayableContract(&transaction)
	if err != nil {
		return nil, err
	}
	if isSendingValueToNonPayableContract {
		return parser.parseSendingValueToNonPayableContract(&transaction)
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

func (parser *IndexedTransactionParser) isSendingValueToNonPayableContract(transaction *IndexedTransaction) (bool, error) {
	hasData := len(transaction.Data) > 0
	isStatusFail := transaction.Status == TransactionStatusFail
	isMistakenlyConsideredRegularTransaction := transaction.Type == TransactionTypeRegular
	if hasData || !isStatusFail && isMistakenlyConsideredRegularTransaction {
		return false, nil
	}

	receiverPubKey, err := parser.pubkeyConverter.Decode(transaction.Receiver)
	if err != nil {
		return false, err
	}

	isReceiverSmartContract := core.IsSmartContractAddress(receiverPubKey)
	return isReceiverSmartContract, nil
}

func (parser *IndexedTransactionParser) parseSendingValueToNonPayableContract(transaction *IndexedTransaction) ([]Operation, error) {
	operations := make([]Operation, 0)

	fee := parsers.MultiplyUint64(parser.minGasLimit, transaction.GasPrice)

	operations = append(operations, Operation{
		Type:        OperationTypeFee,
		Subtype:     OperationSubtypeFeeOfInvalidTransaction,
		Status:      OperationStatusSuccess,
		Address:     transaction.Sender,
		AmountValue: fee.String(),
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
