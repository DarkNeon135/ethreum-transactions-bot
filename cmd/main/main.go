package main

import (
	"EthereumScanner/internal/file"
	"EthereumScanner/internal/log"
	"EthereumScanner/pkg/ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"math/big"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

func main() {

	log.Init("eth", true)

	eth := new(ethereum.Ethereum)
	// if mode 0 - you should insert providers list, because program will swap provider list
	// if mode blank - it will always use first element of the array

	clientMode, err := strconv.Atoi(os.Args[0])
	if err != nil {
		log.Error.Fatal("mode value parsing failed. Error: %s", err)
	}

	providerList := []string{
		//TODO Paste here url/urls_sets/file_path
		//It depends from your mode
	}

	startBlock := int64(14308900) // start block
	//int64(13916167)
	endBlock := int64(15916168) // end block

	eth.Client, err = ethereum.ConnectToEthereum(providerList[0])
	if err != nil {
		log.Error.Fatal(err)
	}
	routinesCount := runtime.NumCPU() * 2

	eth.TransactionCh = make(chan []ethereum.Transaction, routinesCount)

	eth.ErrorCh = make(chan error, routinesCount)

	abiErc20File, err := file.ReadFile("/assets/abi/ERC20/ERC20.abi")
	if err != nil {
		log.Error.Fatal(err)
	}

	eth.Erc20Contract, err = abi.JSON(strings.NewReader(string(abiErc20File)))
	if err != nil {
		log.Error.Fatal(err)
	}

	wg := sync.WaitGroup{}

	transactionArr := make([]ethereum.Transaction, 0)

	addressCounter := 0

	blockCounter := 0

	providerCounter := 1

	addressMap := make(map[string]int)

	addressMap, err = file.RestoreAddressesList("assets/addressList.txt", &addressCounter)

	for i := startBlock; i < endBlock; {

		wg.Add(1)
		go eth.ReceiveBlocksInfo(&wg, big.NewInt(i))
		blockCounter++

		if blockCounter%20 == 0 && clientMode == 0 {
			eth.Client, err = ethereum.ConnectToEthereum(providerList[providerCounter])
			if err != nil {
				log.Error.Fatal(err)
			}
			providerCounter++
			if providerCounter == len(providerList) {
				providerCounter = 0
			}
		}

		if blockCounter%routinesCount == 0 || i == endBlock-1 {

			wg.Wait()

			if len(eth.ErrorCh) != 0 {

				log.Error.Fatal(<-eth.ErrorCh)
			}

			txChLen := len(eth.TransactionCh)
			for j := 0; j < txChLen; j++ {
				transactions := <-eth.TransactionCh
				for _, value := range transactions {
					transactionArr = append(transactionArr, value)
				}
			}
			if addressMap, err = file.WriteTransactionToTxt(addressMap, transactionArr, &addressCounter); err != nil {
				log.Error.Fatal(err)
			}
			transactionArr = []ethereum.Transaction{}

			if i == endBlock-1 {
				log.Warning.Println("Program was successfully finished!")
				break
			}
		}
	}
}
