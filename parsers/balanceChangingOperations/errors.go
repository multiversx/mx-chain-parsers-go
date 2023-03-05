package balanceChangingOperations

import (
	"errors"
)

var errCannotParseTransfer = errors.New("cannot parse transfer")
var errNilPubkeyConverter = errors.New("nil pubkey converter")
var errBadMinGasLimit = errors.New("bad min gas limit")
