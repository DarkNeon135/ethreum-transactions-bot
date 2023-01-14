package file

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

func ReadFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file failed. Error: %s", err)
	}
	defer file.Close()

	result, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading from file failed. Error: %s", err)
	}
	return result, nil
}

func RestoreAddressesList(filePath string, addressCounter *int) (map[string]int, error) {
	addressMap := make(map[string]int)

	addressListArr, err := ReadFile(filePath)
	addressList := string(addressListArr)
	if err != nil {
		return addressMap, fmt.Errorf("file %s reading failed. Error: %s", filePath, err)
	}
	if addressListArr == nil {
		return addressMap, nil
	}

	fileStr := ""
	txId := 0
	for i := 0; i < len(addressList); i++ {
		if addressList[i] == '\t' {
			id := fileStr
			txId, err = strconv.Atoi(id)
			if err != nil {
				return nil, fmt.Errorf("parse address id to int failed. Error: %s", err)
			}
			fileStr = ""
			continue
		}
		if addressList[i] == '\n' {
			addressMap[fileStr] = txId
			fileStr = ""
			continue
		}
		fileStr += string(addressList[i])
	}
	maximumCounter := 0

	for _, id := range addressMap {
		if id > maximumCounter {
			maximumCounter = id
		} else {
			continue
		}
	}

	*addressCounter = maximumCounter + 1

	err = MoveFile(filePath, "assets_backup/addressList.txt")
	if err != nil {
		return nil, fmt.Errorf("file %s moving failed. Error: %s", filePath, err)
	}

	return addressMap, nil
}
