package balanceChangingOperations

import (
	"errors"
)

var errCannotParseTransfer = errors.New("cannot parse transfer")
var errNilPubkeyConverter = errors.New("nil pubkey converter")
var errNilCallArgsParser = errors.New("nil call args parser")
var errBadMinGasLimit = errors.New("bad min gas limit")
var errBadGasLimitPerByte = errors.New("bad gas limit per byte")
