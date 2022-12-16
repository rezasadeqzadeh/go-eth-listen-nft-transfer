package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	pullEvent()
}

func pullEvent() {
	client, err := ethclient.Dial("wss://api.avax-test.network/ext/bc/C/ws")
	if err != nil {
		log.Fatal(err)
	}
	var contractAddresses []common.Address

	addressFile, err := os.Open("contracts.txt")
	if err != nil {
		log.Fatal(err)
	}
	bufRead := bufio.NewScanner(addressFile)
	for bufRead.Scan() {
		contractAddresses = append(contractAddresses, common.HexToAddress(bufRead.Text()))
	}
	fmt.Println("Number of contracts addresses: ", len(contractAddresses))

	// Keck256 Topic : 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
	//Transfer (index_topic_1 address from, index_topic_2 address to, index_topic_3 uint256 tokenId)
	topic := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	topics := [][]common.Hash{
		{topic},
	}

	lastFetchedBlock := 0
	for {
		//get current block number
		currentBlock, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		if lastFetchedBlock == 0 {
			lastFetchedBlock = int(currentBlock) - 1
		}
		fmt.Printf("------------------ Query block range: %d - %d -----------------\n", lastFetchedBlock, currentBlock)

		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(lastFetchedBlock)),
			ToBlock:   big.NewInt(int64(currentBlock)),
			Addresses: contractAddresses,
			Topics:    topics,
		}

		//query logs
		logs, err := client.FilterLogs(context.Background(), query)
		if err != nil {
			fmt.Println(err)
		}

		for _, vLog := range logs {
			fmt.Printf("Block: %d Contract: %s TxHash: %s \n", vLog.BlockNumber, vLog.Address, vLog.TxHash)
		}
		lastFetchedBlock = int(currentBlock) + 1
		time.Sleep(10 * time.Second)
	}
}
