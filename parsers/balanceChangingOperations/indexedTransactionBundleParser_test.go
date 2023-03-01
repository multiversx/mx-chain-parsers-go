package balanceChangingOperations

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	genesisTime   = uint64(1648551600)
	roundDuration = uint64(6)
	apiUrl        = "https://devnet-api.multiversx.com"
)

type balanceRecord struct {
	Timestamp uint64 `json:"timestamp"`
	Balance   string `json:"balance"`
}

func TestFoo(t *testing.T) {
	numTransfers := 1000
	numBalanceRecords := 2000

	parser, err := NewIndexedTransactionBundleParser(IndexedTransactionBundleParserConfig{})
	require.Nil(t, err)

	addresses := []string{
		// Frank
		"erd1kdl46yctawygtwg2k462307dmz2v55c605737dp3zkxh04sct7asqylhyv",
	}

	for _, address := range addresses {
		transfersDescending := fetchTransfers(address, numTransfers)
		balanceRecordsDescending := fetchBalanceRecords(address, numBalanceRecords)

		startingRound := decideStartingRound(transfersDescending)
		startingBalance := findBalanceAtRound(balanceRecordsDescending, startingRound)
		computedBalance := big.NewInt(0).Set(startingBalance)
		fmt.Println("Starting round:", startingRound, "starting balance:", startingBalance)

		for i := len(transfersDescending) - 1; i > 0; i-- {
			txHash := transfersDescending[i].Hash
			round := transfersDescending[i].Round

			if round < startingRound {
				continue
			}

			operations, err := parser.ParseBundle(transfersDescending[i])
			require.Nil(t, err)

			for _, operation := range operations {
				if operation.Address != address {
					continue
				}

				amount := stringToBigInt(operation.AmountValue)

				if operation.Direction == OperationDirectionCredit {
					computedBalance.Add(computedBalance, amount)
				} else {
					computedBalance.Sub(computedBalance, amount)
				}
			}

			actualBalance := findBalanceAtRound(balanceRecordsDescending, round)
			delta := big.NewInt(0).Sub(actualBalance, computedBalance)

			fmt.Println("Round:", round, "computed:", computedBalance, "actual:", actualBalance, "delta:", delta, "txHash:", txHash)
		}
	}
}

func decideStartingRound(transfersDescending []IndexedTransactionBundle) uint64 {
	desiredRoundsGap := uint64(10)
	previousRound := transfersDescending[len(transfersDescending)-1].Round

	for i := len(transfersDescending) - 1; i > 0; i-- {
		if transfersDescending[i].Round > previousRound+desiredRoundsGap {
			return transfersDescending[i].Round - 1
		}

		previousRound = transfersDescending[i].Round
	}

	panic("could not decide on starting round")
}

func findBalanceAtRound(balanceRecordsDescending []balanceRecord, round uint64) *big.Int {
	timestamp := roundToTimestamp(round)

	for i := 0; i < len(balanceRecordsDescending); i++ {
		if balanceRecordsDescending[i].Timestamp <= timestamp {
			return stringToBigInt(balanceRecordsDescending[i].Balance)
		}
	}

	panic(fmt.Sprintf("could not find balance at round %d", round))
}

func roundToTimestamp(round uint64) uint64 {
	return genesisTime + round*roundDuration
}

func fetchTransfers(address string, numItems int) []IndexedTransactionBundle {
	url := fmt.Sprintf("%s/accounts/%s/transfers?size=%d", apiUrl, address, numItems)
	var items []IndexedTransactionBundle
	fetchData(url, &items)
	return items
}

func fetchBalanceRecords(address string, numItems int) []balanceRecord {
	url := fmt.Sprintf("%s/accounts/%s/history?size=%d", apiUrl, address, numItems)
	var items []balanceRecord
	fetchData(url, &items)
	return items
}

func fetchData(url string, data interface{}) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		panic(err)
	}
}

func stringToBigInt(value string) *big.Int {
	result := big.NewInt(0)
	_, _ = result.SetString(value, 10)
	return result
}
