package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
)

// Intel HEX レコードを生成
func createIntelHexRecord(byteCount int, address int, recordType int, data []byte) string {
	checksum := byteCount + (address >> 8) + (address & 0xFF) + recordType

	// データ部
	dataStr := ""
	for _, b := range data {
		dataStr += fmt.Sprintf("%02X", b)
		checksum += int(b)
	}

	// チェックサム計算
	checksum = (256 - (checksum & 0xFF)) & 0xFF

	// HEXレコードフォーマット（アドレスを16進数で明示）
	return fmt.Sprintf(":%02X%04X%02X%s%02X", byteCount, address, recordType, dataStr, checksum)
}

// 拡張リニアアドレスレコードの生成（16進数フォーマット）
func createExtendedLinearAddressRecord(upperAddress int) string {
	data := []byte{byte(upperAddress >> 8), byte(upperAddress & 0xFF)}
	return createIntelHexRecord(2, 0, 4, data)
}

// データレコードの生成（最大0x20バイトで分割）
func createDataRecords(startAddr int, data []byte) []string {
	var records []string
	currentUpperAddr := 0
	offset := 0

	for offset < len(data) {
		// 32バイト (0x20) 以下で分割
		chunkSize := 0x20
		if offset+chunkSize > len(data) {
			chunkSize = len(data) - offset
		}

		// 現在のアドレス
		currentAddr := startAddr + offset
		upperAddr := currentAddr >> 16
		lowerAddr := currentAddr & 0xFFFF

		// 拡張アドレスが変わった場合、拡張リニアアドレスレコードを追加
		if upperAddr != currentUpperAddr {
			records = append(records, createExtendedLinearAddressRecord(upperAddr))
			currentUpperAddr = upperAddr
		}

		// データレコードを追加（アドレスを16進数に統一）
		records = append(records, createIntelHexRecord(chunkSize, lowerAddr, 0, data[offset:offset+chunkSize]))

		offset += chunkSize
	}

	// EOFレコードを追加
	records = append(records, ":00000001FF")

	return records
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go <start_address_hex> <hex_data>")
		os.Exit(1)
	}

	// 開始アドレスを16進数としてパース
	startAddr, err := strconv.ParseUint(os.Args[1], 16, 32)
	if err != nil {
		fmt.Println("Invalid start address")
		os.Exit(1)
	}

	// 16進数データをバイト配列に変換
	data, err := hex.DecodeString(os.Args[2])
	if err != nil {
		fmt.Println("Invalid hex data")
		os.Exit(1)
	}

	// Intel HEXレコードの生成と出力
	records := createDataRecords(int(startAddr), data)
	for _, record := range records {
		fmt.Println(record)
	}

}
