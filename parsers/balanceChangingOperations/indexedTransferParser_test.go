package balanceChangingOperations

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"path"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	genesisTime        = uint64(1648551600)
	roundDuration      = uint64(6)
	apiUrl             = "https://devnet-api.multiversx.com"
	log                = logger.GetOrCreate("balanceChangingOperationsTest")
	pubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, log)
)

type balanceRecord struct {
	Timestamp uint64    `json:"timestamp"`
	Time      time.Time `json:"time"`
	Balance   string    `json:"balance"`
}

func TestBalanceReconciliation(t *testing.T) {
	numTransfers := 1000
	numBalanceRecords := 2500

	parser, err := NewIndexedTransferParser(IndexedTransferParserArgs{
		PubkeyConverter: pubKeyConverter,
		MinGasLimit:     50000,
		GasLimitPerByte: 1500,
	})
	require.Nil(t, err)

	addresses := []string{
		// Alice
		"erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th",
		// Bob
		"erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx",
		// Frank
		"erd1kdl46yctawygtwg2k462307dmz2v55c605737dp3zkxh04sct7asqylhyv",
		// Grace
		"erd1r69gk66fmedhhcg24g2c5kn2f2a5k4kvpr6jfw67dn2lyydd8cfswy6ede",
	}

	for _, address := range addresses {
		reportBuilder := strings.Builder{}
		reportFile := path.Join("testdata/output", fmt.Sprintf("report_%s.txt", address))

		fmt.Println("Testing address:", address)

		transfers := fetchTransfers(address, numTransfers)
		dumpAsJson(fmt.Sprintf("transfers_%s", address), transfers)

		balanceRecords := fetchBalanceRecords(address, numBalanceRecords)
		dumpAsJson(fmt.Sprintf("balanceRecords_%s", address), balanceRecords)

		startingTransfer, startingBalance := decideStartingTransfer(transfers, balanceRecords)
		computedBalance := big.NewInt(0).Set(startingBalance)
		balanceDelta := big.NewInt(0)

		reportBuilder.WriteString(fmt.Sprintf("Starting balance: %s\n", startingBalance))
		reportBuilder.WriteString("\n")

		for i := startingTransfer; i < len(transfers); i++ {
			transfer := transfers[i]
			txHash := transfer.Hash
			round := transfer.Round
			timestamp := transfer.Timestamp
			time := time.Unix(int64(timestamp), 0).UTC()

			reportBuilder.WriteString("\n")
			reportBuilder.WriteString(fmt.Sprintf("Round: %d (%v)\n", round, time))

			if transfer.isSmartContractResult() {
				reportBuilder.WriteString(fmt.Sprintf("Smart Contract Result: %s\n", txHash))
			} else {
				reportBuilder.WriteString(fmt.Sprintf("Transfer: %s\n", txHash))
			}

			operations, err := parser.ParseTransfer(transfer)
			require.Nil(t, err)

			for _, operation := range operations {
				if operation.Address != address {
					continue
				}
				if operation.Status != OperationStatusSuccess {
					continue
				}

				amount := stringToBigInt(operation.AmountValue)

				if operation.Direction == OperationDirectionCredit {
					computedBalance.Add(computedBalance, amount)
					reportBuilder.WriteString(fmt.Sprintf("\tCREDIT:\t + %s (%s)\n", amount, operation.Type))
				} else {
					computedBalance.Sub(computedBalance, amount)
					reportBuilder.WriteString(fmt.Sprintf("\tDEBIT:\t - %s (%s)\n", amount, operation.Type))
				}
			}

			reportBuilder.WriteString(fmt.Sprintf("\t> Computed balance: %s\n", computedBalance))

			if !transfer.isSmartContractResult() {
				actualBalance, actualBalanceTime := findBalanceAtRound(balanceRecords, round)
				balanceDelta = big.NewInt(0).Sub(actualBalance, computedBalance)

				reportBuilder.WriteString(fmt.Sprintf("\t> Actual balance: %s (%v)\n", actualBalance, actualBalanceTime))
				reportBuilder.WriteString(fmt.Sprintf("\t> Delta: %s\n", balanceDelta))
			}
		}

		ioutil.WriteFile(reportFile, []byte(reportBuilder.String()), 0644)
		assert.True(t, big.NewInt(0).Cmp(balanceDelta) == 0, "balance delta is not zero, check report: %s", reportFile)
	}
}

func decideStartingTransfer(transfers []IndexedTransfer, balanceRecords []balanceRecord) (int, *big.Int) {
	desiredGapInSeconds := uint64(60)

	for i := 0; i < len(transfers); i++ {
		transfer := transfers[i]

		for j := 1; j < len(balanceRecords)-1; j++ {
			balanceRecord := balanceRecords[j]

			if balanceRecord.Timestamp == transfer.Timestamp {
				previousBalanceChangeIsFarEnough := balanceRecords[j-1].Timestamp < transfer.Timestamp-desiredGapInSeconds
				nextBalanceChangeIsFarEnough := balanceRecords[j+1].Timestamp > transfer.Timestamp+desiredGapInSeconds
				balance := stringToBigInt(balanceRecords[j-1].Balance)

				if previousBalanceChangeIsFarEnough && nextBalanceChangeIsFarEnough {
					return i, balance
				}
			}
		}
	}

	panic("could not decide on starting transfer")
}

func findBalanceAtRound(balanceRecords []balanceRecord, round uint64) (*big.Int, time.Time) {
	timestamp := roundToTimestamp(round)

	for i := 0; i < len(balanceRecords); i++ {
		recordTimestamp := balanceRecords[i].Timestamp
		recordTime := balanceRecords[i].Time

		if recordTimestamp >= timestamp {
			return stringToBigInt(balanceRecords[i].Balance), recordTime
		}
	}

	panic(fmt.Sprintf("could not find balance at round %d", round))
}

func roundToTimestamp(round uint64) uint64 {
	return genesisTime + round*roundDuration
}

func fetchTransfers(address string, numItems int) []IndexedTransfer {
	url := fmt.Sprintf("%s/accounts/%s/transfers?size=%d", apiUrl, address, numItems)
	var items []IndexedTransfer
	fetchData(url, &items)

	// Sort items by timestamp, ascending
	sort.Slice(items, func(i, j int) bool {
		return items[i].Timestamp < items[j].Timestamp
	})

	// Ignore initial SCRs
	firstRegularTransferIndex := -1

	for i := 0; i < len(items); i++ {
		if !items[i].isSmartContractResult() {
			firstRegularTransferIndex = i
			break
		}
	}

	if firstRegularTransferIndex == -1 {
		panic("no regular transfers found")
	}

	return items[firstRegularTransferIndex:]
}

func fetchBalanceRecords(address string, numItems int) []balanceRecord {
	url := fmt.Sprintf("%s/accounts/%s/history?size=%d", apiUrl, address, numItems)
	var items []balanceRecord
	fetchData(url, &items)

	// Sort items by timestamp, ascending
	sort.Slice(items, func(i, j int) bool {
		return items[i].Timestamp < items[j].Timestamp
	})

	for i := 0; i < len(items); i++ {
		items[i].Time = time.Unix(int64(items[i].Timestamp), 0).UTC()
	}

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

func dumpAsJson(name string, data interface{}) {
	filepath := path.Join("testdata", "output", fmt.Sprintf("%s.json", name))

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(filepath, jsonBytes, 0644)
	if err != nil {
		panic(err)
	}
}
