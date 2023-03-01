package parsers

import (
	"encoding/json"
	"math/big"
	"strings"
)

type dynamicMap map[string]interface{}

func ObjectToMap(value interface{}) (dynamicMap, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	var result dynamicMap
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func MapToObject(obj dynamicMap, value interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, value)
	if err != nil {
		return err
	}

	return nil
}

func IsZeroAmount(amount string) bool {
	if amount == "" {
		return true
	}

	value, ok := big.NewInt(0).SetString(amount, 10)
	if ok {
		return value.Sign() == 0
	}

	return false
}

func IsNonZeroAmount(amount string) bool {
	return !IsZeroAmount(amount)
}

func getMagnitudeOfAmount(amount string) string {
	return strings.Trim(amount, "-")
}

func multiplyUint64(a uint64, b uint64) *big.Int {
	return big.NewInt(0).Mul(big.NewInt(0).SetUint64(a), big.NewInt(0).SetUint64(b))
}

func addBigInt(a *big.Int, b *big.Int) *big.Int {
	return big.NewInt(0).Add(a, b)
}
