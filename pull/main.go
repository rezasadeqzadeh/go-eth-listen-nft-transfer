package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
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

	abi721File, err := os.Open("ERC712-ABI.json")
	if err != nil {
		log.Fatal(err)
	}
	abi721Content, err := ioutil.ReadAll(abi721File)
	if err != nil {
		log.Fatal(err)
	}
	abi721, err := abi.JSON(strings.NewReader(string(abi721Content)))

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
	transferTopic := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	//TransferSingle (index_topic_1 address operator, index_topic_2 address from, index_topic_3 address to, uint256 id, uint256 value)
	transferSingleTopic := common.HexToHash("0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62")

	topics := [][]common.Hash{
		{transferTopic, transferSingleTopic},
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
			if len(vLog.Topics) != 4 { //is not ERC721
				continue
			}
			funcSignature := common.HexToAddress(vLog.Topics[0].Hex())
			var from, to common.Address
			var tokenID int64

			switch funcSignature.Hex() {
			case transferTopic.Hex():
				from = common.HexToAddress(vLog.Topics[1].Hex())
				to = common.HexToAddress(vLog.Topics[2].Hex())
				tokenID, err = HexToInt64(vLog.Topics[3].Hex())
			case transferSingleTopic.Hex():
				from = common.HexToAddress(vLog.Topics[2].Hex())
				to = common.HexToAddress(vLog.Topics[3].Hex())
				tokenInfo := struct {
					ID    int64 `json:"id"`
					Value int64 `json:"value"`
				}{}
				abi721.UnpackIntoInterface(&tokenInfo, "TransferSingle", vLog.Data)
				tokenID = tokenInfo.ID
			}

			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("Block: %d Contract: %s TxHash: %s Frm: %s To: %s TokenId: %d \n", vLog.BlockNumber, vLog.Address, vLog.TxHash, from, to, tokenID)
		}
		lastFetchedBlock = int(currentBlock) + 1
		time.Sleep(10 * time.Second)
	}
}

func HexToInt64(hex string) (int64, error) {
	hex = strings.Replace(hex, "0x", "", -1)
	hex = strings.Replace(hex, "0X", "", -1)
	return strconv.ParseInt(hex, 16, 64)
}
