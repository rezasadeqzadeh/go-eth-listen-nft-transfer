package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("content.txt")
	if err != nil {
		log.Fatal(err)
	}
	buf := bufio.NewScanner(file)
	toWrite := make(map[string]bool)
	for buf.Scan() {
		line := buf.Text()
		if !strings.Contains(line, "Contract:") {
			continue
		}

		parts := strings.Split(line, "Contract:")
		fmt.Println(parts[1])
		toWrite[parts[1]] = true
	}
	fileToWrite, err := os.Create("../contracts.txt")
	if err != nil {
		log.Fatal(err)
	}
	bufWrite := bufio.NewWriter(fileToWrite)
	for k, _ := range toWrite {
		bufWrite.WriteString(k + "\n")
	}
	bufWrite.Flush()
	fileToWrite.Close()
}
