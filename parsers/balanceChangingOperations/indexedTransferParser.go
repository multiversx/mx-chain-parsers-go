package balanceChangingOperations

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-parsers-go/parsers"
	vmcommonParsers "github.com/multiversx/mx-chain-vm-common-go/parsers"
)

type IndexedTransferParserArgs struct {
	PubkeyConverter PubkeyConverter
	MinGasLimit     uint64
	GasLimitPerByte uint64
}

type IndexedTransferParser struct {
	pubkeyConverter PubkeyConverter
	minGasLimit     uint64
	gasLimitPerByte uint64
	callArgsParser  CallArgsParser
}

// NewIndexedTransferParser creates a new IndexedTransferParser
func NewIndexedTransferParser(args IndexedTransferParserArgs) (*IndexedTransferParser, error) {
	if check.IfNil(args.PubkeyConverter) {
		return nil, errNilPubkeyConverter
	}
	if args.MinGasLimit == 0 {
		return nil, errBadMinGasLimit
	}

	return &IndexedTransferParser{
		pubkeyConverter: args.PubkeyConverter,
		minGasLimit:     args.MinGasLimit,
		gasLimitPerByte: args.GasLimitPerByte,
		// This is not passed as a dependency
		callArgsParser: vmcommonParsers.NewCallArgsParser(),
	}, nil
}

// ParseTransfer parses a transaction into a list of balance-changing operations
func (parser *IndexedTransferParser) ParseTransfer(transfer IndexedTransfer) ([]Operation, error) {
	if parser.isStakingRewards(&transfer) {
		return parser.parseStakingRewards(&transfer)
	}
	if transfer.isInvalid() {
		return parser.parseInvalidTransfer(&transfer)
	}
	if transfer.isSmartContractResult() {
		return parser.parseSmartContractResult(&transfer)
	}

	isSendingValueToNonPayableContract, err := parser.isSendingValueToNonPayableContract(&transfer)
	if err != nil {
		return nil, err
	}
	if isSendingValueToNonPayableContract {
		return parser.parseSendingValueToNonPayableContract(&transfer)
	}

	return parser.parseRegularTransfer(&transfer)
}

func (parser *IndexedTransferParser) isStakingRewards(transfer *IndexedTransfer) bool {
	isFromMetachain := transfer.SenderShard == core.MetachainShardId
	isNonZero := parsers.IsNonZeroAmount(transfer.Value)
	hasNoData := len(transfer.Data) == 0

	return isFromMetachain && isNonZero && hasNoData
}

func (parser *IndexedTransferParser) parseStakingRewards(transfer *IndexedTransfer) ([]Operation, error) {
	return []Operation{
		{
			Type:        OperationTypeReward,
			Subtype:     OperationSubtypeStakingRewards,
			Status:      OperationStatusSuccess,
			Address:     transfer.Receiver,
			AmountValue: transfer.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionCredit,
		},
	}, nil
}

func (parser *IndexedTransferParser) parseInvalidTransfer(transfer *IndexedTransfer) ([]Operation, error) {
	operations := make([]Operation, 0)

	operations = append(operations, Operation{
		Type:        OperationTypeFee,
		Subtype:     OperationSubtypeFeeOfInvalidTransaction,
		Status:      OperationStatusSuccess,
		Address:     transfer.Sender,
		AmountValue: transfer.Fee,
		AmountType:  AmountTypeNative,
		Direction:   OperationDirectionDebit,
	})

	if parsers.IsNonZeroAmount(transfer.Value) {
		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusFailure,
			Address:     transfer.Sender,
			AmountValue: transfer.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionDebit,
		})

		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusFailure,
			Address:     transfer.Receiver,
			AmountValue: transfer.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionCredit,
		})
	}

	return operations, nil
}

func (parser *IndexedTransferParser) parseSmartContractResult(transfer *IndexedTransfer) ([]Operation, error) {
	operations := make([]Operation, 0)

	if parsers.IsNonZeroAmount(transfer.Value) {
		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusSuccess,
			Address:     transfer.Sender,
			AmountValue: transfer.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionDebit,
		})

		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusSuccess,
			Address:     transfer.Receiver,
			AmountValue: transfer.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionCredit,
		})
	}

	return operations, nil
}

func (parser *IndexedTransferParser) parseRegularTransfer(transfer *IndexedTransfer) ([]Operation, error) {
	operations := make([]Operation, 0)

	operations = append(operations, Operation{
		Type:        OperationTypeFee,
		Subtype:     OperationSubtypeFeeRegular,
		Status:      OperationStatusSuccess,
		Address:     transfer.Sender,
		AmountValue: transfer.Fee,
		AmountType:  AmountTypeNative,
		Direction:   OperationDirectionDebit,
	})

	if parsers.IsNonZeroAmount(transfer.Value) {
		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusSuccess,
			Address:     transfer.Sender,
			AmountValue: transfer.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionDebit,
		})

		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusSuccess,
			Address:     transfer.Receiver,
			AmountValue: transfer.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionCredit,
		})
	}

	return operations, nil
}

func (parser *IndexedTransferParser) isSendingValueToNonPayableContract(transfer *IndexedTransfer) (bool, error) {
	isStatusFail := transfer.Status == TransferStatusFail
	isMistakenlyConsideredRegularTransaction := transfer.Type == TransferTypeRegular
	if !isStatusFail && isMistakenlyConsideredRegularTransaction {
		return false, nil
	}

	receiverPubKey, err := parser.pubkeyConverter.Decode(transfer.Receiver)
	if err != nil {
		return false, err
	}

	isReceiverSmartContract := core.IsSmartContractAddress(receiverPubKey)
	if !isReceiverSmartContract {
		return false, nil
	}

	_, _, err = parser.callArgsParser.ParseData(string(transfer.Data))
	if err == nil {
		return false, nil
	}

	return true, nil
}

func (parser *IndexedTransferParser) parseSendingValueToNonPayableContract(transfer *IndexedTransfer) ([]Operation, error) {
	operations := make([]Operation, 0)

	gasLimit := parser.minGasLimit + parser.gasLimitPerByte*uint64(len(transfer.Data))
	fee := parsers.MultiplyUint64(gasLimit, transfer.GasPrice)

	operations = append(operations, Operation{
		Type:        OperationTypeFee,
		Subtype:     OperationSubtypeFeeOfInvalidTransaction,
		Status:      OperationStatusSuccess,
		Address:     transfer.Sender,
		AmountValue: fee.String(),
		AmountType:  AmountTypeNative,
		Direction:   OperationDirectionDebit,
	})

	if parsers.IsNonZeroAmount(transfer.Value) {
		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusFailure,
			Address:     transfer.Sender,
			AmountValue: transfer.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionDebit,
		})

		operations = append(operations, Operation{
			Type:        OperationTypeTransfer,
			Subtype:     OperationSubtypeTransferNative,
			Status:      OperationStatusFailure,
			Address:     transfer.Receiver,
			AmountValue: transfer.Value,
			AmountType:  AmountTypeNative,
			Direction:   OperationDirectionCredit,
		})
	}

	return operations, nil
}
