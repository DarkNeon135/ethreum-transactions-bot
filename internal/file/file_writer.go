package file

import (
	"EthereumScanner/pkg/ethereum"
	"fmt"
	"os"
	"strconv"
)

func WriteTransactionToTxt(addressMap map[string]int, transactionArr []ethereum.Transaction, addressCounter *int) (map[string]int, error) {
	transactionFile, err := os.OpenFile("assets/transactions.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open transactions file failed. Error: %s", err)
	}
	defer transactionFile.Close()

	for _, transaction := range transactionArr {
		addressesArr := make([]addresses, 0)

		gasPaid := strconv.FormatFloat(transaction.GasPaid, 'f', -1, 64)
		txValue := strconv.FormatFloat(transaction.TxValue, 'f', -1, 64)

		if _, isExist := addressMap[transaction.SenderAddress]; !isExist {
			addressMap[transaction.SenderAddress] = *addressCounter

			addressesArr = append(addressesArr, addresses{
				id:      *addressCounter,
				address: transaction.SenderAddress,
			})
			*addressCounter++

		}

		if _, isExist := addressMap[transaction.RecipientAddress]; !isExist {
			addressMap[transaction.RecipientAddress] = *addressCounter

			addressesArr = append(addressesArr, addresses{
				id:      *addressCounter,
				address: transaction.RecipientAddress,
			})
			*addressCounter++
		}
		if err = writeAddressesToTxt(addressesArr); err != nil {
			return nil, err
		}

		_, err = transactionFile.WriteString(fmt.Sprintf("%d\t%d\t%s\t%s\t%s\t%s\n",
			addressMap[transaction.SenderAddress], addressMap[transaction.RecipientAddress], txValue, gasPaid, transaction.TxHash, transaction.CryptoName))
		if err != nil {
			return nil, fmt.Errorf("write to transactions file failed. Error: %s", err)
		}

	}
	return addressMap, nil
}

type addresses struct {
	id      int
	address string
}

func writeAddressesToTxt(addressMap []addresses) error {
	addressFile, err := os.OpenFile("assets/addressList.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open addressList file failed. Error: %s", err)
	}

	defer addressFile.Close()

	for _, address := range addressMap {
		_, err = addressFile.WriteString(fmt.Sprintf("%d\t%s\n", address.id, address.address))
		if err != nil {
			return fmt.Errorf("write to address file failed. Error: %s", err)
		}
	}

	return nil
}
