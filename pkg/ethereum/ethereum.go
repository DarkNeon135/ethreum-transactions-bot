package ethereum

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
	"sync"
)

const gweiPrice = 0.000000001

type Transaction struct {
	TxHash           string
	GasPaid          float64
	SenderAddress    string
	RecipientAddress string
	TxValue          float64
	CryptoName       string
}

type Ethereum struct {
	Client        *ethclient.Client
	Erc20Contract abi.ABI
	TransactionCh chan []Transaction
	ErrorCh       chan error
	BlockNumber   uint64
}

func ConnectToEthereum(providerUrl string) (*ethclient.Client, error) {
	client, err := ethclient.Dial(providerUrl)
	if err != nil {

		return nil, fmt.Errorf("connecting to ethereum network failed. Error: %s", err)
	}
	return client, nil
}

func (eth *Ethereum) ReceiveBlocksInfo(wg *sync.WaitGroup, blockNumber *big.Int) {
	defer wg.Done()

	blockInfo, err := eth.Client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		eth.ErrorCh <- fmt.Errorf("retiving block info failed. BlockNumber: %d. Error: %s", blockNumber.Uint64(), err)
		return
	}
	eth.BlockNumber = blockNumber.Uint64()
	transactionArr := make([]Transaction, 0)
	for _, tx := range blockInfo.Transactions() {
		transaction, err := eth.ReceiveTransactionInfo(blockInfo, tx)
		if err != nil {
			eth.ErrorCh <- err
			return
		}
		if transaction.TxHash == "" {
			continue
		}
		transactionArr = append(transactionArr, transaction)
	}

	eth.TransactionCh <- transactionArr

}

func (eth *Ethereum) ReceiveTransactionInfo(blockInfo *types.Block, tx *types.Transaction) (Transaction, error) {

	recepientAddress := ""
	contractAddress := ""

	if len(tx.Data()) > 3 {
		data, err := eth.RetrieveSmartContractInfo(tx.Data())
		if err != nil {
			return Transaction{}, nil
		}
		address, isOk := data["to"].(common.Address)
		if isOk != true {
			return Transaction{}, nil
		}

		recepientAddress = address.String()
		contractAddress = tx.To().String()
	}
	if contractAddress == "" {
		contractAddress = "ETH"
	}

	receipt, err := eth.Client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return Transaction{}, fmt.Errorf("receiving transaction details failed. BlockNumber: %d. Error: %s", blockInfo.NumberU64(), err)
	}

	txDetails, err := retrieveTransactionDetails(tx, blockInfo, receipt)
	if err != nil {
		return Transaction{}, err
	}

	senderAddress := txDetails.senderAddress

	if recepientAddress == "" {
		recepientAddress = tx.To().String()
	}

	return Transaction{
		TxHash:           tx.Hash().String(),
		GasPaid:          txDetails.paidGas,
		RecipientAddress: recepientAddress,
		TxValue:          txDetails.txValue,
		SenderAddress:    senderAddress,
		CryptoName:       contractAddress,
	}, nil

}

type transactionDetails struct {
	senderAddress string
	paidGas       float64
	txnSavings    float64
	txValue       float64
}

func retrieveTransactionDetails(tx *types.Transaction, blockInfo *types.Block, receipt *types.Receipt) (transactionDetails, error) {

	gasUsed := float64(receipt.GasUsed) * gweiPrice
	gasPrice := float64(tx.GasPrice().Uint64()) * gweiPrice
	gasTip := tx.GasTipCap().Uint64()

	var err error

	messageTx := types.Message{}

	txDetails := transactionDetails{}
	if tx.Type() == 0 {
		txDetails.paidGas = gasUsed * gasPrice

		messageTx, err = tx.AsMessage(types.NewEIP155Signer(tx.ChainId()), blockInfo.BaseFee())
		if err != nil {
			return transactionDetails{}, fmt.Errorf("receiving transaction failed. Error: %s", err)
		}

	} else if tx.Type() == 2 {
		txDetails.paidGas = gasUsed * ((float64(blockInfo.BaseFee().Uint64() + gasTip)) * gweiPrice)

		messageTx, err = tx.AsMessage(types.NewLondonSigner(tx.ChainId()), blockInfo.BaseFee())
		if err != nil {
			return transactionDetails{}, fmt.Errorf("receiving transaction failed. Error: %s", err)
		}

	} else {
		return transactionDetails{}, nil
	}

	txDetails.senderAddress = messageTx.From().String()

	txDetails.txValue, _ = weiToEther(tx.Value()).Float64()

	return txDetails, nil
}

func weiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether))
}

func (eth *Ethereum) RetrieveSmartContractInfo(txData []byte) (map[string]interface{}, error) {

	methodSigData := txData[:4]
	method, err := eth.Erc20Contract.MethodById(methodSigData)
	if err != nil {
		return nil, err
	}
	if method.Name != "transfer" {
		return nil, err
	}

	inputsSigData := txData[4:]

	inputsMap := make(map[string]interface{})
	if err = method.Inputs.UnpackIntoMap(inputsMap, inputsSigData); err != nil {
		return nil, err
	}
	return inputsMap, nil
}
