package main

import (
	"EthereumScanner/etherum"
	"EthereumScanner/internal/file"
	"EthereumScanner/internal/log"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"math/big"
	"runtime"
	"strings"
	"sync"
)

func main() {

	log.Init("eth", true)

	eth := new(etherum.Ethereum)
	var err error
	providerList := []string{
		//"https://cloudflare-eth.com",
		"https://mainnet.infura.io/v3/48a78f0616194f4396de44215bf77316",
		"https://mainnet.infura.io/v3/baafcab30663439c8e608cbf5d0ab2bc",
		"https://mainnet.infura.io/v3/8e4bd74d5cd24e7d9eb48dff4ad32080",
		"https://mainnet.infura.io/v3/cacfd1a0b90e4d918e623ffb7802c56a",
		"https://mainnet.infura.io/v3/62ace203b9044b6dae4cb7cf3cc1a179",
		"https://mainnet.infura.io/v3/31fb4afb7b4e4a8490616abe5452d943",
		"https://mainnet.infura.io/v3/00bcd0b671f14a2997e92e24212c6eee",
		"https://mainnet.infura.io/v3/7c2b38cd006d4896aa2d36e0c6f703ef",
		"https://mainnet.infura.io/v3/23af5c68ac4148619d07f131d39c6f85",
		"https://mainnet.infura.io/v3/0e1a6d6db9f84e1c8121b0720a1f411c",
		"https://mainnet.infura.io/v3/4d17541366ae43fc89f8e6e35ede0259",
		"https://mainnet.infura.io/v3/f8d090df1ceb464c8c35a719d62108ea",
		"https://mainnet.infura.io/v3/97eba21df1d74db9890e58f3ba3c9d94",
		"https://mainnet.infura.io/v3/330da592d903418ba8dba331a532df0f",
		"https://mainnet.infura.io/v3/b05339cf1f49451eb68d745aefce976f",
		"https://mainnet.infura.io/v3/01c184a65a2c4af9b21cb07c3952c73a",
		"https://mainnet.infura.io/v3/e20135eb30874ffb8f56dab2b4de49a6",
	}
	eth.Client, err = etherum.ConnectToEthereum(providerList[0])
	if err != nil {
		log.Error.Fatal(err)
	}
	routinesCount := runtime.NumCPU() * 2

	eth.TransactionCh = make(chan []etherum.Transaction, routinesCount)

	eth.ErrorCh = make(chan error, routinesCount)

	abiErc20File, err := file.ReadFile("abi/ERC20/ERC20.abi")
	if err != nil {
		log.Error.Fatal(err)
	}

	eth.Erc20Contract, err = abi.JSON(strings.NewReader(string(abiErc20File)))
	if err != nil {
		log.Error.Fatal(err)
	}

	startBlock := int64(14308900) //int64(13916167)
	endBlock := int64(15916168)
	wg := sync.WaitGroup{}

	transactionArr := make([]etherum.Transaction, 0)

	addressCounter := 0

	blockCounter := 0

	providerCounter := 1

	addressMap := make(map[string]int)

	addressMap, err = file.RestoreAddressesList("assets/addressList.txt", &addressCounter)

	for i := startBlock; i < endBlock; {
		if i == 14371200 {
			i = 14371200 + 1000
		}
		wg.Add(1)
		go eth.ReceiveBlocksInfo(&wg, big.NewInt(i))
		blockCounter++
		i += 50

		if blockCounter%20 == 0 {
			eth.Client, err = etherum.ConnectToEthereum(providerList[providerCounter])
			if err != nil {
				log.Error.Fatal(err)
			}
			providerCounter++
			if providerCounter == len(providerList) {
				providerCounter = 0
			}
		}

		if blockCounter%routinesCount == 0 || i == endBlock-1 || blockCounter == 10000 {

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
			transactionArr = []etherum.Transaction{}

			if blockCounter == 10000 {
				break
			}
		}
	}
}
