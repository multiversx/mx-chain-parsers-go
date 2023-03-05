package balanceChangingOperations

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-parsers-go/parsers"
	vmcommonParsers "github.com/multiversx/mx-chain-vm-common-go/parsers"
)

type IndexedTransactionParserArgs struct {
	PubkeyConverter PubkeyConverter
	MinGasLimit     uint64
	GasLimitPerByte uint64
}

type IndexedTransactionParser struct {
	pubkeyConverter PubkeyConverter
	minGasLimit     uint64
	gasLimitPerByte uint64
	callArgsParser  CallArgsParser
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
		gasLimitPerByte: args.GasLimitPerByte,
		// This is not passed as a dependency
		callArgsParser: vmcommonParsers.NewCallArgsParser(),
	}, nil
}

// ParseTransaction parses a transaction into a list of balance-changing operations
func (parser *IndexedTransactionParser) ParseTransaction(transaction IndexedTransaction) ([]Operation, error) {
	if parser.isStakingRewards(&transaction) {
		return parser.parseStakingRewardsTransaction(&transaction)
	}
	if transaction.isInvalidTransaction() {
		return parser.parseInvalidTransaction(&transaction)
	}
	if transaction.isSmartContractResult() {
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
	isStatusFail := transaction.Status == TransactionStatusFail
	isMistakenlyConsideredRegularTransaction := transaction.Type == TransactionTypeRegular
	if !isStatusFail && isMistakenlyConsideredRegularTransaction {
		return false, nil
	}

	receiverPubKey, err := parser.pubkeyConverter.Decode(transaction.Receiver)
	if err != nil {
		return false, err
	}

	isReceiverSmartContract := core.IsSmartContractAddress(receiverPubKey)
	if !isReceiverSmartContract {
		return false, nil
	}

	_, _, err = parser.callArgsParser.ParseData(string(transaction.Data))
	if err == nil {
		return false, nil
	}

	return true, nil
}

func (parser *IndexedTransactionParser) parseSendingValueToNonPayableContract(transaction *IndexedTransaction) ([]Operation, error) {
	operations := make([]Operation, 0)

	gasLimit := parser.minGasLimit + parser.gasLimitPerByte*uint64(len(transaction.Data))
	fee := parsers.MultiplyUint64(gasLimit, transaction.GasPrice)

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
