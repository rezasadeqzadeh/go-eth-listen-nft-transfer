package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	//pullEvent()
	pushEvent()
}

func pushEvent() {
	client, err := ethclient.Dial("wss://api.avax-test.network/ext/bc/C/ws")
	if err != nil {
		log.Fatal(err)
	}

	// Keck256 Topic : 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
	//Transfer (index_topic_1 address from, index_topic_2 address to, index_topic_3 uint256 tokenId)
	//topic := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	//topics := [][]common.Hash{
	//	{topic},
	//}

	query := ethereum.FilterQuery{
		// Topics:    topics,
		// Addresses: addresses,
	}
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("listening to the network events")
	contractCollections := make(map[string]bool)
	for {
		select {
		case err = <-sub.Err():
			fmt.Println(err)
		case eventLog := <-logs:
			if err != nil {
				log.Println("Failed to unpack")
				continue
			}
			if len(eventLog.Topics) > 1 {
				from := common.BytesToAddress(eventLog.Topics[1].Bytes())
				log.Println("From", from)
			}
			if len(eventLog.Topics) > 2 {
				to := common.BytesToAddress(eventLog.Topics[2].Bytes())
				log.Println("To", to)
			}
			log.Println("Contract:", eventLog.Address)
			log.Println("-----------------------------------")

			if !contractCollections[eventLog.Address.String()] {
				contractCollections[eventLog.Address.String()] = true
			}
			fmt.Println("num contracts:", len(contractCollections))
			if len(contractCollections) >= 500 {
				f, err := os.Create("/tmp/addrs")
				if err != nil {
					log.Fatal(err)
				}
				for k, _ := range contractCollections {
					fmt.Fprintln(f, k)
				}
				f.Close()
				close(logs)
			}
		}
	}
}

func queryLogs() {
	client, err := ethclient.Dial("wss://api.avax-test.network/ext/bc/C/ws")
	if err != nil {
		log.Fatal(err)
	}
	ABI712, err := ioutil.ReadFile("./ERC712-ABI.json")
	if err != nil {
		log.Fatal(err)
	}
	contract712Abi, err := abi.JSON(strings.NewReader(string(ABI712)))
	if err != nil {
		log.Fatal(err)
	}

	address := []common.Address{}
	address = append(address, common.HexToAddress("0xc6EEb811851bfd242c5d65D5B8A3445e466d0321"))
	allLogs := []types.Log{}
	currentBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for {
		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(currentBlock) - 2048),
			ToBlock:   big.NewInt(int64(currentBlock)),
			Addresses: address,
		}
		logs, err := client.FilterLogs(context.Background(), query)
		for _, currentLog := range logs {
			var event map[string]interface{}
			err = contract712Abi.UnpackIntoMap(event, "Transfer", currentLog.Data)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(event)
		}
		allLogs = append(allLogs, logs...)
		if err != nil {
			log.Fatal(err)
		}
		currentBlock = currentBlock - 2048
	}
}
